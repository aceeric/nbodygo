ROOT := $(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))

.PHONY : all
all: proto client server

.PHONY : proto
proto:
	protoc --proto_path=$(ROOT)/cmd/nbodygrpc/ $(ROOT)/cmd/nbodygrpc/nbodyservice.proto \
	  --go_out=plugins=grpc:$(ROOT)/cmd/nbodygrpc

.PHONY : client
client:
	go build -o $(ROOT)/bin/client $(ROOT)/cmd/client

.PHONY : server
server:
	go build -o $(ROOT)/bin/server $(ROOT)/cmd/server

.PHONY : help
help:
	echo "$$HELPTEXT"

ifndef VERBOSE
.SILENT:
endif

.PHONY : print-%
print-%: ; $(info $* is a $(flavor $*) variable set to [$($*)]) @true

export HELPTEXT
define HELPTEXT

This Make file builds bin/client and bin/server relative to the project root. Options are a) run from within
project root, or, b) use the -C make arg if running from outside project root. This Make file assumes the
necessary dependencies (go, protoc) are already installed. The Make file doesn't do any dependency checking,
it just runs the build each time.

Targets:

all      In order, runs: proto, client, server
proto    Runs the 'protoc' protobuf compiler (must be in the PATH) to compile the nbody gRPC service
         protobuf file into Go code. See the cmd/nbodygrpc directory.
client   Runs go build on the cmd/client directory and creates executable bin/client
server   Runs go build on the cmd/server directory and creates executable bin/server
help     Prints this help
print-%  Prints the value of a Make variable. E.g. 'make print-ROOT'

The Make file runs silent unless you provide a VERBOSE arg or variable. E.g.:

make VERBOSE=1
endef
