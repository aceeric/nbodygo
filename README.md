# nbodygo

A project to learn Go. A port (in progress) of https://github.com/aceeric/nbodyjava
 
Rough WIP. Many many TODOs...

sudo apt install xorg-dev libgl1-mesa-dev libopenal1 libopenal-dev \
 libvorbis0a libvorbis-dev libvorbisfile3

bazel run //:gazelle -- update-repos -from_file=go.mod

bazel build not working

but this works:
go build -o $(pwd) ./...
go build -o $(pwd) ./cmd/server

TODO

- gRPC client
- instrumentation
- Bazel or Make
- enums
- logging
- tests
- todos
- build documentation (esp. steps for G3N)
- readme cleanup
- go docs
grpc
https://grpc.io/docs/quickstart/go/
```
$ export GO111MODULE=on
$ go get google.golang.org/grpc@v1.28.1
$ go get github.com/golang/protobuf/protoc-gen-go
$ ls -l $(go env GOPATH)/bin/protoc-gen-go 
-rwxr-xr-x 1 eace eace 8862644 Apr 21 09:42 /home/eace/go/bin/protoc-gen-go
$ export PATH="$PATH:$(go env GOPATH)/bin"
$ pwd
/home/eace/go/nbodygo
$ protoc --proto_path=cmd/nbodygrpc/ cmd/nbodygrpc/nbodyservice.proto --go_out=plugins=grpc:cmd/nbodygrpc
```

