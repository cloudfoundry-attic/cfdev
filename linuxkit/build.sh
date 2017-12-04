#!/bin/bash -e

linuxkit pkg build -hash dev pkg/garden-runc

moby build -name garden -format iso-efi \
   base.yml \
   garden.yml
