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
compiled_dir="${script_dir}/compiled-cache"
iso_file="${script_dir}/cf-oss-deps.iso"

rm -rf "${iso_file}"
rm -rf "${download_list}"

mkdir -p "${download_dir}" "${compiled_dir}"
touch "${download_list}"

# Collect download items
manifest_bosh=$(docker run pivotal/cf-oss cat /etc/bosh/director.yml)
stemcell_url=$(echo "${manifest_bosh}" | bosh int - --path /resource_pools/name=vms/stemcell/url)
#stemcell_version="$(echo "${stemcell_url}" | grep -o 'v=\(\d*\.\d*\)' | cut -d '=' -f 2)"
echo "${manifest_bosh}" | bosh int - --path /releases | grep url | awk '{print $2}' >> "${download_list}"
echo "${stemcell_url}" >> "${download_list}"

manifest_cf=$(docker run pivotal/cf-oss cat /etc/cf/deployment.yml)
stemcell_version=$(echo "$manifest_cf" | bosh interpolate - --path /stemcells/0/version)
echo "$manifest_cf" | bosh int - --path /releases | grep url | awk '{print $2}' >> "${download_list}"
echo "https://bosh.io/d/stemcells/bosh-warden-boshlite-ubuntu-trusty-go_agent?v=$stemcell_version" >> "${download_list}"

iso_dir=$(mktemp -d)

# Place the 'workspace' container image
"${script_dir}/../images/cf-oss/build.sh"
cid=$(docker run -d pivotal/cf-oss sleep infinity)
docker export "$cid" > "${iso_dir}/workspace.tar"
docker kill "$cid"
docker rm "$cid"

wget --no-http-keep-alive -P "${download_dir}" -c -i "${download_list}"

for url in $(cat "${download_list}"); do
 file="${download_dir}/$(basename $url)"

 if ! is_release "${file}"; then
   cp "${file}" "${iso_dir}"
   continue
 fi

 # Already compiled
 if ! has_uncompiled_packages "${file}"; then
   cp "${file}" "${iso_dir}"
   continue
 fi

 compiled_release="${compiled_dir}/$(basename $url)"

 if [ ! -f "${compiled_release}" ]; then
   "${script_dir}/compile-release.sh" "${file}" "${stemcell_version}" "${compiled_release}"
 fi

 cp "${compiled_release}" "${iso_dir}"

done


mkisofs -V cf-oss-deps -R -o "${iso_file}" "${iso_dir}"
