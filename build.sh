#!/bin/sh

rm -f ./grpcdemo/grpcdemo.pb.go
rm -f ./server/server
rm -f ./client/client

protoc -I grpcdemo/ grpcdemo/grpcdemo.proto --go_out=plugins=grpc:grpcdemo

cd couchstore
go build
cd ..

cd server
go build
cd ..

cd client
go build
cd ..


# dependencies:
# https://github.com/couchbase/gocb.git  >> go get github.com/couchbase/gocb
