#!/bin/bash

set -e
# Currently for acceptance we symlink files

STORY_ID="$1"

function fetch() {
    set -e
    rm -rf $1
    wget -O $1 https://s3.amazonaws.com/pcfdev-development/stories/"$STORY_ID"/"$1"
    chmod +x $1
}

# Circle CI build
# Revision 3a00025a275f2654651bf11dad7dda3d64ed9da9
fetch hyperkit

# Build from https://github.com/pcfdev-forks/linuxkit/tree/qcow2
fetch linuxkit

# Built from https://github.com/mirage/ocaml-qcow
# Revision @ 50a7b6612543259850a06582b07650232a36f73c
# Clear your $HOME/.opam directory
# See https://travis-ci.org/mirage/ocaml-qcow/jobs/316643348#L815
# for the commands being invoked by the .travis-opam.sh script
#
# opam init
# opam remote add extra0 https://github.com/djs55/opam-repository.git#io-page
# opam pin add qcow . -n
# opam pin add qcow-tool . -n
# opam depext -u qcow-tool
# opam install qcow-tool "-v"
fetch qcow-tool

# See vpnkit repo for circle ci instructions
# Revision 75434cdd2c2c7c3be257f07f3b7c1a91eca27225
fetch vpnkit



