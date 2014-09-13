package key

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

func Kvkeys(c *cli.Context) {
	prefix := c.Args().First()
	keys, _, _ := kv().Keys(prefix, "", nil)

	for _, key := range keys {
		fmt.Printf("%s\n", key)
	}
}

func KvDelTree(c *cli.Context) {
	prefix := c.Args().First()
	_, err := kv().DeleteTree(prefix, nil)
	if err != nil {
		fmt.Printf("err: %s\n", err)
		return
	}

	fmt.Printf("deleted %s\n", prefix)
}

func Kvlist(c *cli.Context) {
	prefix := c.Args().First()
	pairs, _, _ := kv().List(prefix, nil)

	for _, pair := range pairs {
		fmt.Printf("key: %s, value: %s\n", pair.Key, string(pair.Value))
	}
}

func Kvget(c *cli.Context) {
	key := c.Args().First()

	pair, _, err := kv().Get(key, nil)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return
	}

	if pair == nil {
		fmt.Printf("couldn't find key '%s'", key)
		return
	}

	fmt.Printf("key: %s, value: %s\n", pair.Key, string(pair.Value))
}
