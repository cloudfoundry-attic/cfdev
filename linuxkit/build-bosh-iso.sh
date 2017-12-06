#!/bin/bash

set -e

script_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

$script_dir/../images/deploy-bosh/build.sh

manifest=$(docker run pivotal/deploy-bosh cat /etc/bosh/director.yml)
tmpdir=$(mktemp -d)
trap "{ rm -rf $tmpdir; }" EXIT

rm -rf bosh-deps.iso

pushd "$tmpdir"

echo "$manifest" | bosh int - --path /releases | grep url | awk '{print $2}' > downloads.txt
echo "$manifest" | bosh int - --path /resource_pools/name=vms/stemcell/url >> downloads.txt

mkdir iso

cid=$(docker run -d pivotal/deploy-bosh /bin/sleep 1h)
docker export "$cid" > iso/deploy-bosh.tar
docker kill "$cid"
docker rm "$cid"

wget -c -P iso -i downloads.txt
mkisofs -V bosh-deps -R -o "$script_dir/bosh-deps.iso" iso/*

popd
