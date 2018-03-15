#!/usr/bin/env bash

set -ex

cfdev="/Users/pivotal/workspace/cfdev"
dir="$( cd "$( dirname "$0" )" && pwd )"
cfdev="$dir"/../../..
cache_dir="$HOME"/.cfdev/cache

export GOPATH="$cfdev"
pkg="code.cloudfoundry.org/cfdev/config"

export GOOS=darwin
export GOARCH=amd64

go build \
  -ldflags \
    "-X $pkg.cfdepsUrl=file://$cfdev/output/cf-oss-deps.iso
     -X $pkg.cfdepsMd5=$(md5 $cfdev/output/cf-oss-deps.iso | awk '{ print $4 }')
     -X $pkg.cfdevefiUrl=file://$cfdev/output/cfdev-efi.iso
     -X $pkg.cfdevefiMd5=$(md5 $cfdev/output/cfdev-efi.iso | awk '{ print $4 }')
     -X $pkg.vpnkitUrl=file://$cache_dir/vpnkit
     -X $pkg.vpnkitMd5=$(md5 "$cache_dir"/vpnkit | awk '{ print $4 }')
     -X $pkg.hyperkitUrl=file://$cache_dir/hyperkit
     -X $pkg.hyperkitMd5=$(md5 "$cache_dir"/hyperkit | awk '{ print $4 }')
     -X $pkg.linuxkitUrl=file://$cache_dir/linuxkit
     -X $pkg.linuxkitMd5=$(md5 "$cache_dir"/linuxkit | awk '{ print $4 }')
     -X $pkg.qcowtoolUrl=file://$cache_dir/qcow-tool
     -X $pkg.qcowtoolMd5=$(md5 "$cache_dir"/qcow-tool | awk '{ print $4 }')
     -X $pkg.uefiUrl=file://$cache_dir/UEFI.fd
     -X $pkg.uefiMd5=$(md5 "$cache_dir"/UEFI.fd | awk '{ print $4 }')
     -X $pkg.cfdevdUrl=file://$cache_dir/cfdevd
     -X $pkg.cfdevdMd5=$(md5 "$cache_dir"/cfdevd | awk '{ print $4 }')" \
     code.cloudfoundry.org/cfdev


