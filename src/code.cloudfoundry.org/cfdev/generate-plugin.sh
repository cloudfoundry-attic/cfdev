#!/usr/bin/env bash

cfdev="/Users/pivotal/workspace/cfdev"

export GOPATH=$cfdev
pkg="code.cloudfoundry.org/cfdev/config"

export GOOS=darwin
export GOARCH=amd64

go build \
  -ldflags \
    "-X $pkg.cfdepsUrl=file://$cfdev/output/cf-oss-deps.iso
     -X $pkg.cfdepsMd5=$(md5 ${cfdev}/output/cf-oss-deps.iso | awk '{ print $4 }')
     -X $pkg.cfdevefiUrl=file://$cfdev/output/cfdev-efi.iso
     -X $pkg.cfdevefiMd5=$(md5 ${cfdev}/output/cfdev-efi.iso | awk '{ print $4 }')
     -X $pkg.vpnkitUrl=file://$cfdev/linuxkit/vpnkit
     -X $pkg.vpnkitMd5=$(md5 ${cfdev}/linuxkit/vpnkit | awk '{ print $4 }')
     -X $pkg.hyperkitUrl=file://$cfdev/linuxkit/hyperkit
     -X $pkg.hyperkitMd5=$(md5 ${cfdev}/linuxkit/hyperkit | awk '{ print $4 }')
     -X $pkg.linuxkitUrl=file://$cfdev/linuxkit/linuxkit
     -X $pkg.linuxkitMd5=$(md5 ${cfdev}/linuxkit/linuxkit | awk '{ print $4 }')
     -X $pkg.qcowtoolUrl=file://$cfdev/linuxkit/qcow-tool
     -X $pkg.qcowtoolMd5=$(md5 ${cfdev}/linuxkit/qcow-tool | awk '{ print $4 }')
     -X $pkg.uefiUrl=file://$cfdev/linuxkit/UEFI.fd
     -X $pkg.uefiMd5=$(md5 ${cfdev}/linuxkit/UEFI.fd | awk '{ print $4 }')" \
     code.cloudfoundry.org/cfdev


