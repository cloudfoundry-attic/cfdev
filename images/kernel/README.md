### CF Dev Kernel

We're removed the btrfs quota checks in the kernel since it wasn't working
properly with v4.9.x. See patch 0013.

The motivation to patch the kernel instead the diego release was motivated by
us not wanting to maintain a forked diego BOSH release. This will keep
pulling upstream compiled releases simple.

We've copied the relevant 4.9.x configs and Dockerfile from the linuxkit source 

Source: https://github.com/linuxkit/linuxkit/tree/5a294e5840471fa111afd35daf5d56a92e8c5d3f/kernel

