#!/bin/bash

go version

export GOPATH=~/go
export PATH=$PATH:$GOPATH/bin

echo "go get packages ..."
go get gopkg.in/mgo.v2
go get github.com/petar/GoLLRB/llrb

echo "go build ..."
go build -o bin/botsbox src/main.go

echo "copy configuration ..."
mkdir -p bin/
cp config.json bin/