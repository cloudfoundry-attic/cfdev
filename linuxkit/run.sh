#!/bin/bash

set -e

linuxkit run hyperkit \
	-networking=vpnkit \
	-disk size=20G \
	-disk file=bosh-deps.iso \
	--uefi cfdev-efi.iso

