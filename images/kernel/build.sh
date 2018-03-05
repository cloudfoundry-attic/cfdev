#!/bin/bash

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

docker build \
    --build-arg KERNEL_VERSION=4.9.78 \
    --build-arg KERNEL_SERIES=4.9.x \
    -t cfdev/kernel:4.9.78 \
    "$DIR"
