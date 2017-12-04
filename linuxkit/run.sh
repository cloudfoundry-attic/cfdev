#!/bin/bash

linuxkit run hyperkit \
	-networking=vpnkit \
	-disk size=20G \
	-disk file=bosh-deps.iso \
	--uefi garden-efi.iso

