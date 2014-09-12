build: main.go
	gox -output "bin/{{.Dir}}_{{.OS}}_{{.Arch}}" -os "linux darwin" -arch "amd64"
	docker build -t jmcarbo/consul_servant .

run:
	docker run -d -h node1 --privileged --name node1 jmcarbo/consul_servant /consul_servant
	docker inspect node1
