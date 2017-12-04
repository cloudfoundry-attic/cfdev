#!/bin/bash -e

linuxkit pkg build -hash dev pkg/garden-runc
linuxkit pkg build -hash dev pkg/openssl

moby build -name cfdev -format iso-efi \
   base.yml \
   garden.yml
