#!/bin/sh

export GOPATH=$GOPATH:`pwd`
rm -rf exec
go build -race -o exec fs.go
./exec