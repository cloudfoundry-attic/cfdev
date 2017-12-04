#!/bin/bash

linuxkit run hyperkit \
	-networking=vpnkit \
	-disk size=20G \
	--uefi garden-efi.iso

