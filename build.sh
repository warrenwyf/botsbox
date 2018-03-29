#!/bin/bash

go version

export GOPATH=~/go
export PATH=$PATH:$GOPATH/bin

echo "go get packages ..."
go get github.com/petar/GoLLRB/llrb
go get github.com/tidwall/gjson
go get github.com/PuerkitoBio/goquery
go get github.com/gotk3/gotk3/gtk
go get github.com/mattn/go-sqlite3
go get gopkg.in/mgo.v2

echo "go build ..."
go build -o bin/botsbox src/main.go

echo "copy misc files ..."
mkdir -p bin/misc/
cp misc/* bin/misc/