#!/bin/bash

set -e

script_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

$script_dir/../images/deploy-cf/build.sh

manifest=$(docker run pivotal/deploy-cf cat /etc/cf/deployment.yml)
tmpdir=$(mktemp -d)
trap "{ rm -rf $tmpdir; }" EXIT

rm -rf cf-deps.iso

pushd "$tmpdir"

stemcell_version=$(echo "$manifest" | bosh interpolate - --path /stemcells/0/version)

echo "https://bosh.io/d/stemcells/bosh-warden-boshlite-ubuntu-trusty-go_agent?v=$stemcell_version" > downloads.txt
echo "$manifest" | bosh int - --path /releases | grep url | awk '{print $2}' >> downloads.txt

mkdir iso

cid=$(docker run -d pivotal/deploy-cf /bin/sleep 1h)
docker export "$cid" > iso/deploy-cf.tar
docker kill "$cid"
docker rm "$cid"

wget -c -P iso -i downloads.txt
mkisofs -V cf-deps -R -o "$script_dir/cf-deps.iso" iso/*


popd
