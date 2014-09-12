build: main.go
	gox -output "bin/{{.Dir}}_{{.OS}}_{{.Arch}}" -os "linux darwin" -arch "amd64"
	docker build -t jmcarbo/consul_servant .

start_cluster:
	export NODE1=$(shell docker run -d -h node1 --privileged --name node1 jmcarbo/consul_servant /consul_servant)
	export CIP=$(shell docker inspect --format '{{ .NetworkSettings.IPAddress }}' node1)
	@echo $(CIP)
	export NODE2=$(shell docker run -d -h node2 --privileged --name node2 jmcarbo/consul_servant /consul_servant -server -join @echo $(CIP) )
	export NODE3=$(shell docker run -d -h node3 --privileged --name node3 jmcarbo/consul_servant /consul_servant -join @echo $(CIP))

stop_cluster:
	docker rm -f node1
	docker rm -f node2
	docker rm -f node3

CIP=$(shell docker inspect --format '{{ .NetworkSettings.IPAddress }}' node1)
run_client:
	docker run -ti -e CIP="$(CIP)" --rm --name client1 --privileged jmcarbo/consul_servant bash

push:
	docker push jmcarbo/consul_servant
