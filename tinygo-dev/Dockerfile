
# NOTE: this is needs to be run with the parent directory as the context and paths are relative to that

# FROM tinygo/tinygo:0.9.0
FROM tinygo/tinygo-dev:latest

RUN apt-get update && apt-get install -y curl git

#RUN wget https://dl.google.com/go/go1.13.4.linux-amd64.tar.gz -O /root/go.tar.gz
RUN curl https://dl.google.com/go/go1.13.4.linux-amd64.tar.gz -o /root/go.tar.gz
RUN mkdir -p /opt && cd /opt && tar -xzvf /root/go.tar.gz

#RUN tinygo version
#RUN echo $GOPATH
RUN GOROOT=/opt/go GOPATH=/go GO111MODULE=off /opt/go/bin/go get github.com/vugu/xxhash github.com/vugu/html github.com/vugu/vjson

COPY / /go/src/github.com/vugu/vugu/
#COPY /tinygo-dev/ /go/src/testpgm/

#COPY src/ /go/src/
#COPY main1.go /go/src/wasmtest/main1.go

#CMD ["tinygo", "build", "-o", "/out/tinygo-dev/testpgm.wasm", "-target", "wasm", "testpgm"]

