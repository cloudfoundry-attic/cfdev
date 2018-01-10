#!/bin/bash

set -e
set -u

release_path=$1
stemcell_version=$2
output_file=${3-""}
stemcell_name="bosh-warden-boshlite-ubuntu-trusty-go_agent?v=$stemcell_version"
stemcell_url="https://bosh.io/d/stemcells/$stemcell_name"

# import the stemcell into docker and install ruby
image_id=$(docker images -q "bosh/stemcells:${stemcell_version}")

if [ -z "$image_id" ]; then
    tmpdir=$(mktemp -d)

    pushd "$tmpdir" > /dev/null
      wget --continue "$stemcell_url"
      tar zxvf "${stemcell_name}"
      docker import image "bosh/stemcells:${stemcell_version}-raw"
      cid=$(docker run -d "bosh/stemcells:${stemcell_version}-raw" bash -e -c 'apt-get update && apt-get install -y ruby')
      exit_code=$(docker wait "${cid}")

      if [ "${exit_code}" -ne "0" ]; then
        echo "docker failed to run bosh/stemcells"
        exit 1
      fi

      docker commit "${cid}" "bosh/stemcells:${stemcell_version}"
      docker rm "${cid}"
      docker rmi "bosh/stemcells:${stemcell_version}-raw"
    popd > /dev/null
fi

cid=$(docker run -d -w /staging "bosh/stemcells:${stemcell_version}" sleep infinity)

trap '{ docker rm -f ${cid}; }' EXIT

docker exec "${cid}" mkdir -p /var/vcap/data/compile /staging/compiled_packages
docker cp "${release_path}" "${cid}:/staging/release.tgz"
docker exec "${cid}" tar zxvf release.tgz jobs packages release.MF
docker exec -u root:root "${cid}" ruby -ryaml -rset -e "

release = YAML.load_file('/staging/release.MF')

puts 'Compiling release ' + release['name'] + '...'

Dir.chdir('/staging/packages')

packages = {}

release['packages'].each do |pkg|
   name = pkg['name']

   packages[name] = pkg

   %x[ mkdir -p /var/vcap/data/compile/#{name} ]
   %x[ tar zxf #{name}.tgz -C /var/vcap/data/compile/#{name} ]
end

compiled = Set[]
outstanding = packages.keys.to_set

while !outstanding.empty?
  outstanding.each do |name|
    pkg = packages[name]
    deps = pkg['dependencies'].to_set

    if !compiled.superset?(deps)
      next
    end

    puts 'Compiling package ' + name + '...'
    ENV[\"BOSH_INSTALL_TARGET\"]  = '/var/vcap/packages/' + name
    ENV[\"BOSH_COMPILE_TARGET\"]  = '/var/vcap/data/compile' + name
    ENV[\"BOSH_PACKAGE_NAME\"]    = name
    ENV[\"BOSH_PACKAGE_VERSION\"] = pkg['version']

    %x[ mkdir -p \$BOSH_INSTALL_TARGET ]

    Dir.chdir('/var/vcap/data/compile/' + name)

    %x[ bash -x packaging ]

    outstanding.delete(name)
    compiled.add(name)
  end
end


# Best way to do a deep copy is load
release['compiled_packages'] = release.delete('packages')
release.delete('license') # Seems like compiled releases don't have license info

compiled.each do |name|
   Dir.chdir('/var/vcap/packages/' + name)
   %x[ tar czvf /staging/compiled_packages/#{name}.tgz . ]

   sha = %x[ sha256sum /staging/compiled_packages/#{name}.tgz | cut -d ' ' -f 1 ]
   packages[name]['sha1'] = 'sha256:' + sha.to_s.strip
   packages[name]['stemcell'] = 'ubuntu-trusty/$stemcell_version'
end


Dir.chdir('/staging')
File.open('release.MF', 'w') {|f| f.write release.to_yaml }
%x[ rm -rf packages ]

tar_name = release['name'] + '-' + release['version'] + '-ubuntu-trusty-' + '$stemcell_version'

%x[ chown -R vcap:vcap /staging ]
%x[ tar czvf #{tar_name}.tgz compiled_packages jobs release.MF ]

%x[ rm -rf release.tgz ]
%x[ rm -rf compiled_packages ]
%x[ rm -rf release.MF ]
%x[ rm -rf jobs ]

"

compiled_release_file=$(docker exec -i "${cid}" ls)

if [ -z "${output_file}" ]; then
    output_file=compiled_release_file
fi

docker cp "${cid}:/staging/${compiled_release_file}" "${output_file}"
