#!/bin/sh

protoc -I grpcdemo/ grpcdemo/grpcdemo.proto --go_out=plugins=grpc:grpcdemo