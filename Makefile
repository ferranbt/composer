SHELL := /bin/bash

protoc:
	protoc --go_out=. --go-grpc_out=. ./proto/*.proto
