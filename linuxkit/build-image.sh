#!/bin/bash -e


script_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
"${script_dir}/../images/kernel/build.sh"

linuxkit pkg build -hash dev pkg/bosh-lite-routing
linuxkit pkg build -hash dev pkg/expose-multiple-ports
linuxkit pkg build -hash dev pkg/garden-runc
linuxkit pkg build -hash dev pkg/openssl

linuxkit build \
  -disable-content-trust \
  -name cfdev \
  -format iso-efi \
   base.yml \
   garden.yml
