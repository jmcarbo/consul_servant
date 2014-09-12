Consul Servant is a Consul based cluster manager and orchestrator. It is a single executable that
installs consul in the current directory and starts a new cluster:

```

./consul_servant

```

To add more nodes run the following command in another machine (virtual or metal)

```
./consul_servant -join <firt node ip> [-server]
```

Note that you can add either server or client nodes to the consul mesh.

Now you have an orchestrated consul server with two job queues http://localhost:8500/v1/kv/jobs and 
http://localhost:8500/v1/kv/queues/<node name>. The first queue "jobs" is honoured by any consul node, while
the <node name> queue is honoured only by the named node.

To submit a job just type

```
curl -X PUT -d '{ "Command": "echo hello world" }' http://localhost:8500/v1/kv/jobs/<job id>

or 

curl -X PUT -d '{ "Command": "echo hello world" }' http://localhost:8500/v1/kv/queues/<node name>/<job id>
```

The job will be runned by the first node that gets the job payroll. Note that job id must be unique across the cluster.

Job results can be found at:

```
curl -X GET http://localhost:8500/v1/kv/jdone_jid/jobs/<job id>?raw

or

curl -X GET http://localhost:8500/v1/kv/queues/<node name>/jdone_jid/jobs/<job id>?raw
```

Current extra job parameters are:

```
{ "Command": "command to run", 
  "NoWait": true|false, 
  "Timeout": <timeout in seconds default 60 seconds>,
  "Type": "<shell| ...>",
  "Services": [ {"ID": "service id", "Name": "service Name", ... }]
  }
```

Job with service registration:

```
curl -X PUT -d '{"Command": "docker ps", "Timeout": 3, "Services": [ { "ID": "blu3", "Name": "blu3", "Port": "80", "Check": {"TTL": "23s"}} ] }' http://localhost:8500/v1/kv/jobs/61
```

Thats all for now. I accept further development suggestions. 
