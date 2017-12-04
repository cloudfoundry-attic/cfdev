#!/bin/bash

set -e

current_dir="$PWD"
manifest=$(docker run dprotaso/deploy-bosh cat /etc/bosh/director.yml)
tmpdir=$(mktemp -d)

pushd "$tmpdir"

echo "$manifest" | bosh int - --path /releases | grep url | awk '{print $2}' > downloads.txt
echo "$manifest" | bosh int - --path /resource_pools/name=vms/stemcell/url >> downloads.txt

mkdir iso

wget -c -P iso -i downloads.txt
mkisofs -V bosh-deps -R -o "$current_dir/bosh-deps.iso" iso/*

popd
