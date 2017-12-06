#!/bin/bash

# This still requires Docker For Mac

set -e

rm -rf cfdev-efi-state/

linuxkit run hyperkit \
	-cpus 4 \
	-mem 8192 \
	-networking=vpnkit \
	-disk size=50G \
	-disk file=bosh-deps.iso \
	-disk file=cf-deps.iso \
	--uefi cfdev-efi.iso
