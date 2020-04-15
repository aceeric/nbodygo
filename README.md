# nbodygo
https://nbodygo/internal/pkg

sudo apt install xorg-dev libgl1-mesa-dev libopenal1 libopenal-dev \
 libvorbis0a libvorbis-dev libvorbisfile3

bazel run //:gazelle -- update-repos -from_file=go.mod

bazel build not working

but this works:
go build -o $(pwd) ./...

open requirements:
1. add/remove geo from scene including lightsource
2. Movement of bodies

resolved requirements
ability to create a sphere with one line of code
flycam: WASDQZ / Mouse look
cam speed
engage/disengage keyboard

TODO
body
 - result queue holder
 - computation runner
sim
 - main
 - nbodysim
 - simgen (only one sim for starters)