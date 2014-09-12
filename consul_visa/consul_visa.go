package main

import (
  "os"
  "github.com/codegangsta/cli"
  "github.com/armon/consul-api"
  "github.com/jmcarbo/consul_servant/job"
  "fmt"
  "encoding/json"
  "github.com/nu7hatch/gouuid"
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
    cli.StringFlag{
      Name: "action,a",
      Value: "",
      Usage: "Command to execute",
    },
  }

  startFlags := []cli.Flag {
    cli.StringFlag{
      Name: "command,c",
      Value: "",
      Usage: "Command to execute",
    },
  }

  app.Commands = []cli.Command{
    {
      Name:      "start",
      ShortName: "s",
      Usage:     "Start a job in the cluster",
      SkipFlagParsing: false,
      Flags: startFlags,
      Action: func(c *cli.Context) {
        if c.String("command") != "" {
          id := c.Args().First()
          if id == "" {
            u, _:=uuid.NewV4()
            id = u.String()
          }
          jsn := fmt.Sprintf("{ \"ID\": \"%s\", \"Command\": \"%s\" }", id, c.String("command"))
          p := &consulapi.KVPair{Key: "jobs/"+id, Value: []byte(jsn) }
          _, err := kv.Put(p, nil)
          if err != nil {
            fmt.Printf("Error: %s\n", err)
            panic("Error scheduling job")
          }
          fmt.Printf("scheduled job: %s\n", id)

        } else {
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
            fmt.Printf("scheduled job: %s / %s\n", c.Args().First(), j.ID)
          }
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
