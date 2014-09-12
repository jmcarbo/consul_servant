docker run -d -h node1 --privileged --name node1 jmcarbo/consul_servant /consul_servant
export CIP=$(docker inspect --format '{{ .NetworkSettings.IPAddress }}' node1)
docker run -d -h node2 --privileged --name node2 jmcarbo/consul_servant /consul_servant -server -join $CIP 
docker run -d -h node3 --privileged --name node3 jmcarbo/consul_servant /consul_servant -join $CIP
