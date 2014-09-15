SUBDIRS = consul_servant consul_visa
      
.PHONY: subdirs $(SUBDIRS)
		      
subdirs: $(SUBDIRS)
	
$(SUBDIRS):
	$(MAKE) -C $@

build_docker:
	docker build -t jmcarbo/consul_servant .

start_cluster:
	export NODE1=$(shell docker run -d -h node1 --privileged --name node1 jmcarbo/consul_servant /consul_servant)
	export CIP=$(shell docker inspect --format '{{ .NetworkSettings.IPAddress }}' node1)
	@echo $(CIP)
	sleep 5
	export NODE2=$(shell docker run -d -h node2 --privileged --name node2 jmcarbo/consul_servant /consul_servant -server -join @echo $(CIP) )
	sleep 5
	export NODE3=$(shell docker run -d -h node3 --privileged --name node3 jmcarbo/consul_servant /consul_servant -join @echo $(CIP))

stop_cluster:
	docker rm -f node1
	docker rm -f node2
	docker rm -f node3

CIP=$(shell docker inspect --format '{{ .NetworkSettings.IPAddress }}' node1)
run_client:
	docker run -ti -e CIP="$(CIP)" --rm -h client1 --name client1 --privileged jmcarbo/consul_servant bash

push:
	docker push jmcarbo/consul_servant
