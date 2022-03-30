#!/bin/sh
cd p2pclient
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags "-s -w" -o ../p2pclient.exe main.go
cd -


cd signalserver
go build -ldflags "-s -w" -o ../server main.go
cd -

strip ss cc