#!/bin/bash

# This still requires Docker For Mac

set -ex

while getopts "i:c:" arg; do
  case $arg in
    i) image="$OPTARG"
      ;;
    c) cf_deps_iso="$OPTARG"
      ;;
  esac
done

cache_dir="$HOME"/.cfdev/cache
script_dir="$( cd "$(dirname "$0")" && pwd)"

if [[ -z $image ]]; then
    image="$cache_dir"/cfdev-efi.iso
fi
if [[ -z $cf_deps_iso ]]; then
    cf_deps_iso="$cache_dir"/cf-oss-deps.iso
fi


rm -rf $script_dir/cfdev-efi-state

linuxkit_bin="$cache_dir/linuxkit"
hyperkit_bin="$cache_dir/hyperkit"
vpnkit_bin="$cache_dir/vpnkit"
qcowtool_bin="$cache_dir/qcow-tool"
uefi_fw="$cache_dir/UEFI.fd"

$linuxkit_bin run hyperkit \
    -console-file \
    -state "$script_dir"/cfdev-efi-state \
    -hyperkit "$hyperkit_bin" \
	-cpus 4 \
	-mem 8192 \
	-fw "$uefi_fw" \
	-networking vpnkit \
	-vpnkit "$vpnkit_bin" \
	-disk type=qcow,size=50G,trim=true,qcow-tool=$qcowtool_bin,qcow-onflush=os,qcow-compactafter=262144,qcow-keeperased=262144 \
	-disk file="$cf_deps_iso" \
	--uefi "$image"
