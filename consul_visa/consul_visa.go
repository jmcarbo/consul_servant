package main

import (
  "os"
  "github.com/codegangsta/cli"
  "github.com/armon/consul-api"
  "github.com/jmcarbo/consul_servant/job"
  "fmt"
  "encoding/json"
)

var (
  client *consulapi.Client
  kv *consulapi.KV
  session *consulapi.Session
  agent *consulapi.Agent
  agent_name string
)

func init () {
  client, err := consulapi.NewClient(consulapi.DefaultConfig())
  if err != nil {
    panic(err)
  }

  kv = client.KV()
  session = client.Session()
  agent = client.Agent()
  agent_name, _ = agent.NodeName()
}

func main() {
  app:=cli.NewApp()
  app.Name = "consul_visa"
  app.Usage = "access consul_servant cluster from the command line"
  app.Flags = []cli.Flag {
    cli.StringFlag{
      Name: "lang",
      Value: "english",
      Usage: "language for the greeting",
    },
  }
  app.Commands = []cli.Command{
    {
      Name:      "start",
      ShortName: "s",
      Usage:     "Start a job in the cluster",
      Action: func(c *cli.Context) {
        j, err := job.LoadJobFromFile(c.Args().First())
        if err != nil {
          fmt.Printf("Error reading job file: %s", err) 
        } else {
          p := &consulapi.KVPair{Key: "jobs/"+j.ID, Value: j.Encode() }
          _, err := kv.Put(p, nil)
          if err != nil {
            fmt.Printf("Error: %s\n", err)
            panic("Error scheduling job")
          }
          fmt.Printf("scheduled job: %s\n", c.Args().First())
        }
      },
    },
    {
      Name:      "show",
      ShortName: "w",
      Usage:     "Show job status",
      Action: func(c *cli.Context) {
        p, _, err := kv.Get("jdone_jid/jobs/"+c.Args().First(), nil)
        if err != nil {
          panic(err) 
        }
        if p != nil {
          var dat map[string] interface{}
          if err := json.Unmarshal(p.Value, &dat); err != nil {
            panic(err)
          }
          out, _ := json.MarshalIndent(dat,"", "    ")
          fmt.Printf("Job status: %s\n", string(out))
        } else {
          fmt.Printf("Job not found\n")
        }
      },
    },
  }
  app.Run(os.Args)
}
