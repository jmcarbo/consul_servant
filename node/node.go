package node

import (
	"fmt"
  "github.com/codegangsta/cli"
	"github.com/armon/consul-api"
)


type runStatus int

const (
	runOK runStatus = iota
	runErr
	runExit
)

func client() *consulapi.Client {
	client, _ := consulapi.NewClient(consulapi.DefaultConfig())
	return client
}

func kv() *consulapi.KV {
	return client().KV()
}

func agent() *consulapi.Agent {
	return client().Agent()
}

func FindNode(nodeName string) bool {
	members, err := agent().Members(false)
	if err != nil {
		fmt.Printf("err: can't list members\n")
		return false
	}

	for _, member := range members {
		if member.Name == nodeName {
			return true
		}
	}

	return false
}

func NodeEject(c *cli.Context) {
	nodeName := c.Args().First()

	if FindNode(nodeName) {
		err := agent().ForceLeave(nodeName)
		if err != nil {
			fmt.Printf("err: %s\n", err)
			return
		}
	} else {
		fmt.Printf("err: cound't find node %s\n", nodeName)
	}

	fmt.Printf("removed %s\n", nodeName)
}

func NodeList(c *cli.Context) {
	members, err := agent().Members(false)
	if err != nil {
		fmt.Printf("err: can't list members\n")
		return
	}

	for _, member := range members {
		fmt.Println(member.Name)
	}
}
