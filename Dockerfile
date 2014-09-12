FROM jpetazzo/dind
RUN apt-get update && apt-get install -y wget unzip curl
ADD bin/test_linux_amd64 /consul_servant 
ADD examples.sh /examples.sh
