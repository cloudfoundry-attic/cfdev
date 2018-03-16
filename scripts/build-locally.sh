#!/usr/bin/env bash
echo "GENERATE CLOUD CONFIG"

./generate-cloud-config -c ~/workspace/cf-deployment/

echo "GENERATE CF MANIFEST"

./generate-cf-manifest -c ~/workspace/cf-deployment/

echo "GENERATE CF DEPS TAR"

./build-cf-deps-tar -m $PWD/../output/cf.yml -c $PWD/../output/cloud-config.yml

echo "GENERATE BOSH MANIFEST"

./generate-bosh-manifest -b ~/workspace/bosh-deployment/

echo "GENERATE BOSH TAR"

./build-bosh-deps-tar -m $PWD/../output/bosh.yml

echo "GENERATE CF ISO"

./build-cf-deps-iso -c $PWD/../output/cf.tgz  -b $PWD/../output/bosh.tgz

echo "BUID EFI IMAGE"

./build-image

cd $PWD/../

echo "NOW, PLEASE GENERATE CF PLUGIN VIA: $PWD/src/code.cloudfoundry.org/cfdev/generate-plugin.sh"
