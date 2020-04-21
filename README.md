# nbodygo

A project to learn Go. A port (in progress) of https://github.com/aceeric/nbodyjava
 
Rough WIP. Many many TODOs...

sudo apt install xorg-dev libgl1-mesa-dev libopenal1 libopenal-dev \
 libvorbis0a libvorbis-dev libvorbisfile3

bazel run //:gazelle -- update-repos -from_file=go.mod

bazel build not working

but this works:
go build -o $(pwd) ./...

TODO

- https://developers.google.com/protocol-buffers/docs/reference/go-generated
- gRPC client
- metrics
- Bazel
- logging
- id generator
- tests
- make the code idiomatic
- todos
- etc..

grpc
https://grpc.io/docs/quickstart/go/
go get google.golang.org/grpc@v1.28.1