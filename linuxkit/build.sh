#!/bin/bash -e

linuxkit pkg build -hash dev pkg/garden-runc
linuxkit pkg build -hash dev pkg/openssl

moby build -name garden -format iso-efi \
   base.yml \
   garden.yml
