#!/bin/bash

set -e

function has_uncompiled_packages() {
  tar xOf "$1" release.MF 2> /dev/null | grep -q ^packages
}

function is_release() {
  tar tf "$1" 2> /dev/null | grep -q release.MF
}

script_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
download_dir="${script_dir}/download-cache"
download_list="${download_dir}/downloads.txt"
iso_file="${script_dir}/cf-oss-deps.iso"

rm -rf "${iso_file}"
rm -rf "${download_list}"


if [ -z "$SKIP_PURGE" ]; then
  rm -rf "${download_dir}"
fi

mkdir -p "${download_dir}"
touch "${download_list}"

"${script_dir}/../images/cf-oss/build.sh"

manifest_bosh=$(docker run pivotal/cf-oss cat /etc/bosh/director.yml)
stemcell_url=$(echo "${manifest_bosh}" | bosh int - --path /resource_pools/name=vms/stemcell/url)
#stemcell_version="$(echo "${stemcell_url}" | grep -o 'v=\(\d*\.\d*\)' | cut -d '=' -f 2)"
echo "${manifest_bosh}" | bosh int - --path /releases | grep url | awk '{print $2}' >> "${download_list}"
echo "${stemcell_url}" >> "${download_list}"

manifest_cf=$(docker run pivotal/cf-oss cat /etc/cf/deployment.yml)
stemcell_version=$(echo "$manifest_cf" | bosh interpolate - --path /stemcells/0/version)
echo "$manifest_cf" | bosh int - --path /releases | grep url | awk '{print $2}' >> "${download_list}"
echo "https://bosh.io/d/stemcells/bosh-warden-boshlite-ubuntu-trusty-go_agent?v=$stemcell_version" >> "${download_list}"

cid=$(docker run -d pivotal/cf-oss sleep infinity)
docker export "$cid" > "${download_dir}/workspace.tar"
docker kill "$cid"
docker rm "$cid"

wget --no-http-keep-alive -P "${download_dir}" -c -i "${download_list}"

iso_dir=$(mktemp -d)

for f in "${download_dir}/"*; do
 if ! is_release "$f"; then
   cp "${f}" "${iso_dir}"
   continue
 fi

 if ! has_uncompiled_packages "$f"; then
   cp "${f}" "${iso_dir}"
   continue
 fi

 "${script_dir}/compile-release.sh" "${f}" "${stemcell_version}" "${iso_dir}/$(basename ${f})"
done

mkisofs -V cf-oss-deps -R -o "${iso_file}" "${iso_dir}"
