for i in `seq 1 10`;
do
#    curl -X PUT -d '{"Command":"echo my taylor is rich $i"}' http://localhost:8500/v1/kv/jobs/$i
  consul_visa start -c "echo blabla $i" 
done
#curl -X PUT -d '{ "Command": "wrapdocker", "NoWait": true }' http://localhost:8500/v1/kv/queues/node1/1
#curl -X PUT -d '{"Command": "docker run -d rufus/isawesome" }' http://localhost:8500/v1/kv/jobs/40
#curl -X PUT -d '{"Command": "docker ps" }' http://localhost:8500/v1/kv/jobs/41
#curl http://localhost:8500/v1/kv/jdone_jid/jobs/70?raw
#curl -X PUT -d '{"Command": "docker ps", "Timeout": 3, "Services": [ { "ID": "bla", "Name": "bla", "Check": {"TTL": "23s"}} ] }' http://localhost:8500/v1/kv/jobs/53
#curl http://localhost:8500/v1/catalog/service/bla
#curl http://localhost:8500/v1/agent/checks
#curl -X PUT -d '{"Command": "docker ps", "Timeout": 3, "Services": [ { "ID": "blu3", "Name": "blu3",  "Check": {"TTL": "23s"}} ] }' http://localhost:8500/v1/kv/jobs/61
