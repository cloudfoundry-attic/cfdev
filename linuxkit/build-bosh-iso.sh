#!/bin/bash

set -e

function has_uncompiled_packages() {
  tar xOf "$1" release.MF 2> /dev/null | grep -q ^packages
}

function is_release() {
  tar tf "$1" 2> /dev/null | grep -q release.MF
}

script_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

rm -rf "${script_dir}/bosh-deps.iso"

tmpdir=$(mktemp -d)
trap '{ rm -rf ${tmpdir}; }' EXIT
pushd "${tmpdir}"
mkdir iso

"${script_dir}/../images/deploy-bosh/build.sh"
cid=$(docker run -d pivotal/deploy-bosh /bin/sleep 1h)
docker export "${cid}" > iso/deploy-bosh.tar
docker rm -f "${cid}"

manifest=$(docker run pivotal/deploy-bosh cat /etc/bosh/director.yml)
stemcell_url=$(echo "${manifest}" | bosh int - --path /resource_pools/name=vms/stemcell/url)

echo "${stemcell_url}" > downloads.txt
stemcell_version="$(echo "${stemcell_url}" | grep -o 'v=\(\d*\.\d*\)' | cut -d '=' -f 2)"


echo "${manifest}" | bosh int - --path /releases | grep url | awk '{print $2}' >> downloads.txt

wget -c -P iso -i downloads.txt

for f in iso/*; do
 if ! is_release "$f"; then
   continue
 fi

 if ! has_uncompiled_packages "$f"; then
   continue
 fi

 "${script_dir}/compile-release.sh" "$f" "${stemcell_version}" "$PWD/$f"
done


mkisofs -V bosh-deps -R -o "$script_dir/bosh-deps.iso" iso/*

popd
