FROM jpetazzo/dind
RUN apt-get update && apt-get install -y wget unzip curl
ADD bin/consul_servant_linux_amd64 /consul_servant 
ADD bin/consul_visa_linux_amd64 /consul_visa 
ADD examples.sh /examples.sh
ADD consul_visa/examples /examples
