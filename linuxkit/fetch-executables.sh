#!/bin/bash

# Currently for acceptance we symlink files

rm -rf hyperkit vpnkit linuxkit

wget -O hyperkit https://s3.amazonaws.com/pcfdev-development/hyperkit
wget -O linuxkit https://s3.amazonaws.com/pcfdev-development/linuxkit
wget -O vpnkit   https://s3.amazonaws.com/pcfdev-development/vpnkit

chmod +x hyperkit linuxkit vpnkit

