#!/bin/bash -e


script_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
"${script_dir}/../images/kernel/build.sh"

"${script_dir}"/linuxkit pkg build -hash dev pkg/bosh-lite-routing
"${script_dir}"/linuxkit pkg build -hash dev pkg/expose-multiple-ports
"${script_dir}"/linuxkit pkg build -hash dev pkg/garden-runc
"${script_dir}"/linuxkit pkg build -hash dev pkg/openssl

"${script_dir}"/linuxkit build \
 -disable-content-trust \
 -name cfdev \
 -format iso-efi \
  base.yml \
  garden.yml
