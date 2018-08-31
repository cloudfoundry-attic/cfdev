#!/usr/bin/env bash
set -exo pipefail

cfdev="/Users/pivotal/workspace/cfdev"
dir="$( cd "$( dirname "$0" )" && pwd )"
cfdev="$dir"/../../..
cache_dir="$HOME"/.cfdev/cache
analyticskey="WFz4dVFXZUxN2Y6MzfUHJNWtlgXuOYV2"

export GOPATH="$cfdev"
pkg="code.cloudfoundry.org/cfdev/config"

export GOOS=darwin
export GOARCH=amd64

cfdevd="$PWD"/cfdvd
go build -o $cfdevd code.cloudfoundry.org/cfdev/cfdevd

analyticsd="$PWD"/analytix
analyticsdpkg="main"
go build \
  -o $analyticsd \
  -ldflags \
    "-X $analyticsdpkg.analyticsKey=$analyticskey" \
     code.cloudfoundry.org/cfdev/analyticsd

cfdepsUrl="$cfdev/output/cf-deps.iso"
if [ ! -f "$cfdepsUrl" ]; then
  cfdepsUrl="$cache_dir/cf-deps.iso"
fi
cfdevefiUrl="$cfdev/output/cfdev-efi.iso"
if [ ! -f "$cfdevefiUrl" ]; then
  cfdevefiUrl="$cache_dir/cfdev-efi.iso"
fi

go build \
  -ldflags \
    "-X $pkg.cfdepsUrl=file://$cfdepsUrl
     -X $pkg.cfdepsMd5=$(md5 $cfdepsUrl | awk '{ print $4 }')
     -X $pkg.cfdepsSize=$(wc -c < $cfdepsUrl | tr -d '[:space:]')

     -X $pkg.cfdevefiUrl=file://$cfdevefiUrl
     -X $pkg.cfdevefiMd5=$(md5 $cfdevefiUrl | awk '{ print $4 }')
     -X $pkg.cfdevefiSize=$(wc -c < $cfdevefiUrl | tr -d '[:space:]')

     -X $pkg.vpnkitUrl=file://$cache_dir/vpnkit
     -X $pkg.vpnkitMd5=$(md5 "$cache_dir"/vpnkit | awk '{ print $4 }')
     -X $pkg.vpnkitSize=$(wc -c < "$cache_dir"/vpnkit | tr -d '[:space:]')

     -X $pkg.hyperkitUrl=file://$cache_dir/hyperkit
     -X $pkg.hyperkitMd5=$(md5 "$cache_dir"/hyperkit | awk '{ print $4 }')
     -X $pkg.hyperkitSize=$(wc -c < "$cache_dir"/hyperkit | tr -d '[:space:]')

     -X $pkg.linuxkitUrl=file://$cache_dir/linuxkit
     -X $pkg.linuxkitMd5=$(md5 "$cache_dir"/linuxkit | awk '{ print $4 }')
     -X $pkg.linuxkitSize=$(wc -c < "$cache_dir"/linuxkit | tr -d '[:space:]')

     -X $pkg.qcowtoolUrl=file://$cache_dir/qcow-tool
     -X $pkg.qcowtoolMd5=$(md5 "$cache_dir"/qcow-tool | awk '{ print $4 }')
     -X $pkg.qcowtoolSize=$(wc -c < "$cache_dir"/qcow-tool | tr -d '[:space:]')

     -X $pkg.uefiUrl=file://$cache_dir/UEFI.fd
     -X $pkg.uefiMd5=$(md5 "$cache_dir"/UEFI.fd | awk '{ print $4 }')
     -X $pkg.uefiSize=$(wc -c < "$cache_dir"/UEFI.fd | tr -d '[:space:]')

     -X $pkg.cfdevdUrl=file://$cfdevd
     -X $pkg.cfdevdMd5=$(md5 "$cfdevd" | awk '{ print $4 }')
     -X $pkg.cfdevdSize=$(wc -c < "$cfdevd" | tr -d '[:space:]')

     -X $pkg.analyticsdUrl=file://$analyticsd
     -X $pkg.analyticsdMd5=$(md5 "$analyticsd" | awk '{ print $4 }')
     -X $pkg.analyticsdSize=$(wc -c < "$analyticsd" | tr -d '[:space:]')

     -X $pkg.cliVersion=0.0.$(date +%Y%m%d-%H%M%S)
     -X $pkg.analyticsKey=$analyticskey" \
     code.cloudfoundry.org/cfdev


