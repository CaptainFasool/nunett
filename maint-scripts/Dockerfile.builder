FROM ubuntu:20.04

RUN apt update && DEBIAN_FRONTEND=noninteractive apt install git curl wget libc6 make build-essential dpkg-dev devscripts lintian libsystemd-dev pandoc -y

# Golang install
RUN wget https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
RUN tar -xf go1.21.5.linux-amd64.tar.gz
RUN mv go /usr/local/
RUN ln -s /usr/local/go/bin/go /usr/local/bin/go
RUN ln -s /usr/local/go/bin/gofmt /usr/local/bin/gofmt
RUN rm go1.21.5.linux-amd64.tar.gz