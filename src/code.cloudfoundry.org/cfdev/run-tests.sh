#!/bin/bash

if [[ $EUID -eq 0 ]]; then
   echo "This script must not be run as root"
   exit 1
fi

ginkgo -r -v -skipPackage privileged

pushd acceptance/privileged > /dev/null
  sudo -E ginkgo -r -v privileged
popd > /dev/null
