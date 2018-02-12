#!/bin/bash

set -e

extend_sudo_timeout() {
  while true; do
    sudo -v
    sleep 60
  done
}

disable_sudo() {
    set +e
    sudo -K
}

script_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

echo "The tests require sudo password to be set"
sudo echo "thanks!"
trap disable_sudo EXIT

# We need to extend sudo timeout for our acceptance test to run
extend_sudo_timeout &

cd "$script_dir"

pushd acceptance/privileged > /dev/null
  ginkgo "$@"
popd > /dev/null

# Invalidate sudo credentials
disable_sudo

ginkgo -r -skipPackage privileged "$@"
