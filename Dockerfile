FROM golang:1.16-buster AS base

FROM base AS debug
RUN go get github.com/go-delve/delve/cmd/dlv
RUN mkdir /.cache && chmod -R 777 /.cache
# Add /go/bin to PATH
ENV PATH="/go/bin:${PATH}" 
# Install redis in order to be able to work inside a single container
RUN apt-get update && apt-get install -y redis netcat
RUN chmod -R 777 /go
### Run the Delve debugger ###
COPY /src /src
ENTRYPOINT redis-server --daemonize yes && dlv debug --headless --log-output=debugger --log -l 0.0.0.0:2345 --api-version=2 cmd/main.go

FROM base AS production
RUN mkdir /.cache && chmod -R 777 /.cache
ENV PATH="/go/bin:${PATH}" 
RUN chmod -R 777 /go
COPY /src /src
WORKDIR /src
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o master-finder cmd/main.go  && cp master-finder /usr/local/bin/master-finder
ENTRYPOINT master-finder 