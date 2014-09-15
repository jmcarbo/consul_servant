FROM jpetazzo/dind
RUN apt-get update && apt-get install -y wget unzip curl
RUN mkdir /data
ADD bin/consul_servant_linux_amd64 /consul_servant 
ADD bin/consul_visa_linux_amd64 /consul_visa 
ADD ./consul_servant/examples.sh /examples.sh
ADD ./consul_visa/examples /examples
