build:
	#gox .
	docker build -t jmcarbo/consul_servant .

run:
	docker run -d -h node1 --privileged --name node1 jmcarbo/consul_servant /consul_servant
	docker inspect node1
