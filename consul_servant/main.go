package main

import ( 
  "fmt"
  "runtime"
  "github.com/codeskyblue/go-sh"
  "github.com/armon/consul-api"
  "os"
  "os/exec"
  "time"
  "strings"
  "path"
  "log"
  "encoding/json"
  "io"
  "bytes"
  "sync"
  "io/ioutil"
  "github.com/codegangsta/cli"
)

const (
  execution_timeout = time.Duration(60)
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
}

const default_config = `
{
  "data_dir": "./consul_data",
  "ui_dir": "./consul_ui",
  "client_addr": "127.0.0.1",
  "ports": {
      "dns": 53
  },
  "recursor": "8.8.8.8"
}
`

const consul_install_script = `
rm -f consul;  
unamestr=$(uname);  
if [[ "$unamestr" == 'Linux' ]]; 
then wget -nc -q https://dl.bintray.com/mitchellh/consul/0.4.0_linux_amd64.zip; 
elif [[ "$unamestr" == 'Darwin' ]];  
then wget -nc -q https://dl.bintray.com/mitchellh/consul/0.4.0_darwin_amd64.zip;  
fi; 
unzip -q 0*; rm 0*; 
mkdir -p consul_ui; cd consul_ui; wget -nc -q https://dl.bintray.com/mitchellh/consul/0.4.0_web_ui.zip; 
unzip -q 0*; rm 0*;mv dist/* .
`
func check(e error) {
  if e != nil {
    panic(e)
  }
}

func installConsul(c *cli.Context) {
  cwd, err := os.Getwd()
  if _, err = os.Stat(path.Join(cwd,"consul")); os.IsNotExist(err) {
    log.Printf("Installing consul\n")
    _, err = sh.Command("/bin/bash", "-c", consul_install_script).Output()
    if err != nil {
      panic(err)
    }
  }

  err=os.MkdirAll("./consul_config", 0755)
  check(err)
  err=os.MkdirAll("./consul_data", 0755)
  check(err)
  d1 := []byte(default_config)
  err = ioutil.WriteFile("./consul_config/config.json", d1, 0644)
  check(err)
}

func launchConsul(c *cli.Context) {
  log.Println("Starting consul")
  var cmd *exec.Cmd
  var options []string

  options=append(options, "agent")
  options=append(options, "-data-dir")
  options=append(options, "./consul_data")
  if c.Bool("ui")==true {
    options=append(options,"-ui-dir")
    options=append(options,"./consul_ui")
  }
  if c.String("config-dir")!="" {
    options=append(options,"-config-dir")
    options=append(options,c.String("config-dir"))
  }
  if c.String("join") == "" {
    options=append(options,"-server")
    options=append(options,"-bootstrap")
  } else {
    options=append(options,"-join")
    options=append(options,c.String("join"))

    if c.Bool("server")==true {
      options=append(options,"-server")
    }
  }
  log.Printf("Starting consul with options: %v", options)
  cmd = exec.Command("./consul", options...)
  err := cmd.Start()
  if err != nil {
    panic(err)
  }
  cmd.Wait()
}

type Check struct {
  Script string
  Interval string
  TTL string
}

type Service struct {
  ID string
  Name string
  Tags []string
  Port int
  Check Check
}

type Job struct {
  ID, Name, Command, Output, OutputErrors string
  Type string // default is "shell"
  NoWait bool
  StartTime int64
  EndTime int64
  StartTimeStr string
  EndTimeStr string
  ExecutionNode string
  Timeout time.Duration //Timeout in seconds
  ExitErrors string
  Services []Service
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
              lapsus := execution_timeout
              if job.Timeout > 0 {
                lapsus = job.Timeout
              }
              sss := sh.Command("/bin/bash", "-c", string(job.Command)).SetTimeout(lapsus * time.Second)
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
          for _,s := range job.Services {
            log.Printf("Adding check to service %#v", s.Check)
            var check *consulapi.AgentServiceCheck
            if (s.Check.Script != "") || (s.Check.TTL != "")  {
              check = &consulapi.AgentServiceCheck{ Script: s.Check.Script, Interval: s.Check.Interval, TTL: s.Check.TTL }              
              err=agent.ServiceRegister(&consulapi.AgentServiceRegistration { ID: s.ID, Name: s.Name, Tags: s.Tags, Port: s.Port, Check: check })
            } else {
              err=agent.ServiceRegister(&consulapi.AgentServiceRegistration { ID: s.ID, Name: s.Name, Tags: s.Tags, Port: s.Port  })
            }

            if err!= nil {
              log.Printf("Error creating service %s: %s", s.Name, err)
            } else {
              log.Printf("Service %s created", s.Name)
            }
          }
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
  app:=cli.NewApp()
  app.Name = "consul_servant"
  app.Usage = "consul_servant cluster orchestrator"
  app.Version = "0.0.1"
  //app.Flags = []cli.Flag {
  startFlags := []cli.Flag {
    cli.StringFlag{
      Name: "join,j",
      Value: "",
      Usage: "consul ip agent to join to",
    },
    cli.StringFlag{
      Name: "config-dir",
      Value: "",
      Usage: "pass config-dir option to consul agent",
    },
    cli.BoolFlag{
      Name: "server,s",
      Usage: "Adquire server role",
    },
    cli.BoolFlag{
      Name: "ui",
      Usage: "Add consul ui",
    },
  }

  app.Commands = []cli.Command{
		{
			Name:   "install",
			Usage:  "install consul server",
			Action: installConsul,
		},
    {
      Name:      "start",
      ShortName: "s",
      Usage:     "Start consul_servant orchestator",
      Flags: startFlags,
      Action: func(c *cli.Context) {
        installConsul(c)

        go launchConsul(c)
        time.Sleep(5000 * time.Millisecond)

        client, err := consulapi.NewClient(consulapi.DefaultConfig())
        if err != nil {
          panic(err)
        }

        kv = client.KV()
        session = client.Session()
        agent = client.Agent()
        agent_name, _ = agent.NodeName()
        if agent_name == "" {
          agent_name, _ = os.Hostname()
        }

        log.Printf("Consul Servant connected to %s", agent_name)

        var wg sync.WaitGroup
        wg.Add(1)
        go processQueue("jobs", &wg)
        wg.Add(1)
        go processQueue("queues/"+agent_name, &wg)
        wg.Wait()
      },
    },
  }
  app.Run(os.Args)


}
