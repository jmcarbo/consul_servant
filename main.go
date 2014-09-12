package main

import ( 
  "fmt"
  "runtime"
  "github.com/codeskyblue/go-sh"
  "github.com/armon/consul-api"
  "os"
  "os/exec"
  "time"
  "flag"
  "strings"
  "path"
  "log"
  "encoding/json"
  "io"
  "bytes"
  "sync"
  "io/ioutil"
)

const (
  execution_timeout = 60
)
var (
  join_ip string
  server  bool
  kv *consulapi.KV
  session *consulapi.Session
  agent *consulapi.Agent
  agent_name string
)

func init() {
  _ = runtime.GOMAXPROCS(runtime.NumCPU())

  flag.StringVar(&join_ip, "join", "", "join ip")
  flag.BoolVar(&server, "server", false, "server role")
}

func installConsul() {
  cwd, err := os.Getwd()
  if _, err = os.Stat(path.Join(cwd,"consul")); os.IsNotExist(err) {
    log.Printf("Installing consul\n")
    _, err = sh.Command("/bin/bash", "-c", "rm -f consul; unamestr=`uname`; if [[ \"$unamestr\" == 'Linux' ]]; then wget -q https://dl.bintray.com/mitchellh/consul/0.4.0_linux_amd64.zip; elif [[ \"$unamestr\" == 'Darwin' ]]; then wget -q https://dl.bintray.com/mitchellh/consul/0.4.0_darwin_amd64.zip; fi; unzip -q 0*; rm 0*").Output()
    if err != nil {
      panic(err)
    }
  }
}

func launchConsul() {
  log.Println("Starting consul")
  var cmd *exec.Cmd

  if join_ip == "" {
    cmd = exec.Command("./consul", "agent", "-server", "-bootstrap", "-data-dir", "./data")
  } else {
    if server==true {
      cmd = exec.Command("./consul", "agent", "-server", "-join", join_ip,  "-data-dir", "./data")
    } else {
      cmd = exec.Command("./consul", "agent", "-join", join_ip,  "-data-dir", "./data")
    }
  }
  err := cmd.Start()
  if err != nil {
    panic(err)
  }
  cmd.Wait()
}


type Job struct {
  Name, Command, Output, OutputErrors string
  NoWait bool
  StartTime int64
  EndTime int64
  StartTimeStr string
  EndTimeStr string
  ExecutionNode string
  Timeout int64 //Timeout in seconds
  ExitErrors string
}

func processQueue(qname string, wg *sync.WaitGroup) {
  log.Printf("Start processing queue %s", qname)
  p := &consulapi.KVPair{Key: qname }
  _, err := kv.Put(p, nil)

  pair, _, err := kv.Get(qname, nil)
  if err != nil {
    panic(err)
  }

  modi := pair.ModifyIndex
  for true {

    keys, _, err := kv.List(qname, &consulapi.QueryOptions{ AllowStale: false, RequireConsistent: true, WaitIndex: modi })
    if err != nil {
      fmt.Println(err)
    }

    for _,a := range keys {
      if a.ModifyIndex > modi {
        modi = a.ModifyIndex
      }

      if a.Key == qname {
        continue
      }

      if a.Session != "" {
        continue
      }

      if a.Flags == 99 {
        continue
      }


      ses, _, err := session.CreateNoChecks(nil, nil)
      a.Session = ses

      result, _, err := kv.Acquire(a, nil)
      if err != nil {
        fmt.Println(err)
      }

      if result == false {
        _, err = session.Destroy(ses, nil)
        if err != nil {
          fmt.Println(err)
        }
        continue
      }

      pair, _, err := kv.Get(strings.Replace(a.Key,qname,"jdone_jid/"+qname,1), 
        &consulapi.QueryOptions{ RequireConsistent: true, AllowStale: false })
      if err != nil {
          panic(err)
      }

      if pair != nil {
        _, err = session.Destroy(ses, nil)
        if err != nil {
          fmt.Println(err)
        }
        _, err := kv.Delete(a.Key,nil)
        if err != nil {
          fmt.Println(err)
        }
        continue
      }

      pair, _, err = kv.Get(a.Key, &consulapi.QueryOptions{ RequireConsistent: true, AllowStale: false })
      if err != nil {
          panic(err)
      }

      if pair != nil {
        if pair.ModifyIndex > modi {
          modi = pair.ModifyIndex
        }

        if (ses == pair.Session)  {
          log.Printf("Executing %s **** %s\n", pair.Key, pair.Value)
          job, err := decodeJob(string(pair.Value))
          job.StartTime = time.Now().UnixNano()
          job.StartTimeStr = fmt.Sprintln(time.Now())
          job.ExecutionNode = agent_name 
          if err != nil {
            log.Println("Error decoding job json")
            job.Output = "Error decoding job json"
          } else {
            if job.NoWait == true {
              sss := sh.Command("/bin/bash", "-c", string(job.Command))
              sss.Stdout = ioutil.Discard
              sss.Stderr = ioutil.Discard
              sss.Start()
            } else {
              sss := sh.Command("/bin/bash", "-c", string(job.Command)).SetTimeout(execution_timeout * time.Second)
              out, stderr, err := OutputAll(sss)
              if string(out) != "" {
                log.Printf("Output job %s **** %s\n", pair.Key, string(out))
              }
              if string(stderr) != "" {
                log.Printf("Error job %s **** %s\n", pair.Key, string(stderr))
              }
              if err != nil {
                log.Printf("Exit error %v", err)
                job.ExitErrors = fmt.Sprintf("%v", err)
              }
              job.Output = string(out)
              job.OutputErrors = string(stderr)
            }
          } 
          job.EndTime = time.Now().UnixNano() 
          job.EndTimeStr = fmt.Sprintln(time.Now())
          ori_key := pair.Key
          pair.Key = strings.Replace(pair.Key, qname, "jdone/"+qname, 1) + "/" + agent_name

          en, _ := json.Marshal(job)
          pair.Value = en

          _, err = kv.Put(pair, nil)
          pair.Key = strings.Replace(ori_key, qname, "jdone_jid/"+qname, 1) 
          pair.Session = ses
          result,_, err = kv.Acquire(pair, nil)
          if result==false {
            panic(err)
          }
          _, err = kv.Delete(ori_key,nil)
          if err != nil {
            fmt.Println(err)
          }
          _,_, err = kv.Release(pair,nil)
        }
      }
      _, err = session.Destroy(ses, nil)
      if err != nil {
        fmt.Println(err)
      }
    }

  }

  wg.Done()
}

func decodeJob(payload string) (Job, error) {
  dec := json.NewDecoder(strings.NewReader(payload))
  //for {
    var j Job
    if err := dec.Decode(&j); err == io.EOF {
      //break
      return j, err
    } else if err != nil {
      return j, err
    }
  //}
  return j, nil
}

func OutputAll(s *sh.Session) (out []byte, oerr []byte, err error) {
  oldout := s.Stdout
  olderr := s.Stderr
  defer func() {
      s.Stdout = oldout
      s.Stderr = olderr
  }()
  stdout := bytes.NewBuffer(nil)
  stderr := bytes.NewBuffer(nil)
  s.Stdout = stdout
  s.Stderr = stderr
  err = s.Run()
  out = stdout.Bytes()
  oerr = stderr.Bytes()
  return
}

func main() {
  flag.Parse()

  installConsul()

  go launchConsul()
  time.Sleep(5000 * time.Millisecond)

  client, err := consulapi.NewClient(consulapi.DefaultConfig())
  if err != nil {
    panic(err)
  }

  kv = client.KV()
  session = client.Session()
  agent = client.Agent()
  agent_name, _ = agent.NodeName()

  log.Printf("Consul Servant connected to %s", agent_name)

  var wg sync.WaitGroup
  wg.Add(1)
  go processQueue("jobs", &wg)
  wg.Add(1)
  go processQueue("queues/"+agent_name, &wg)
  wg.Wait()

}
