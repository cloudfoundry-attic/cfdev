#!/bin/bash

set -e

extend_sudo_timeout() {
  while true; do
    sudo -v
    sleep 60
  done
}

disable_sudo() {
    if [ ! -z "${NONPRIV_USER:-}" ] ; then
        (export GOTMPDIR=$(sudo -E su $NONPRIV_USER -c "mktemp -d")
        export GOCACHE=$(sudo -E su $NONPRIV_USER -c "mktemp -d")
        sudo rm -rf $GOPATH/pkg
        mkdir -p $GOPATH/pkg
        sudo chmod 777 $GOPATH/pkg
        trap "sudo rm -rf $GOPATH/pkg $GOTMPDIR $GOCACHE" EXIT
        sudo -E su $NONPRIV_USER -c "$*")
    else
        sudo -E -k "$@"
    fi
}

script_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

echo "The tests require sudo password to be set"
sudo echo "thanks!"

# We need to extend sudo timeout for our acceptance test to run
extend_sudo_timeout &

cd "$script_dir"

pushd acceptance/privileged > /dev/null
    ginkgo "$@"
popd > /dev/null

# Invalidate sudo credentials
disable_sudo ginkgo -r -skipPackage privileged "$@"
