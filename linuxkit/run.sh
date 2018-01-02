#!/bin/bash

# This still requires Docker For Mac

set -e

script_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

rm -rf $script_dir/cfdev-efi-state


linuxkit_bin="$script_dir/linuxkit"
hyperkit_bin="$script_dir/hyperkit"
vpnkit_bin="$script_dir/vpnkit"
qcowtool_bin="$script_dir/qcow-tool"
uefi_fw="$script_dir/UEFI.fd"

$linuxkit_bin run hyperkit \
    -console-file \
    -hyperkit $hyperkit_bin \
	-cpus 4 \
	-mem 8192 \
	-fw $uefi_fw \
	-networking vpnkit \
	-vpnkit $vpnkit_bin \
	-disk type=qcow,size=50G,trim=true,qcow-tool=$qcowtool_bin,qcow-onflush=os,qcow-compactafter=262144,qcow-keeperased=262144 \
	-disk file=bosh-deps.iso \
	-disk file=cf-deps.iso \
	--uefi cfdev-efi.iso
