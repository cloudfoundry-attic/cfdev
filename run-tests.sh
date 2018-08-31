#!/bin/bash

set -e

extend_sudo_timeout() {
  while true; do
    sudo -v
    sleep 60
  done
}

script_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

echo "The tests require sudo password to be set"
sudo echo "thanks!"

# We need to extend sudo timeout for our acceptance test to run
extend_sudo_timeout &

cd "$script_dir"

ginkgo -r -skipPackage privileged "$@"

pushd cfdevd > /dev/null
   ginkgo -v -r "$@"
popd > /dev/null

pushd acceptance/privileged > /dev/null
   ginkgo -v "$@"
popd > /dev/null
