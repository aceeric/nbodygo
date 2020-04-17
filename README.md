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

sim
- performance - currently horrible
- g3n coordinate system reversed vs. JME?
- main -- arg parse, etc
- simgen (only one sim for starters)
- get Bazel working
- make the code idiomatic
- etc..