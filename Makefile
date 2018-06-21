.PHONY: all
all: cfdev

vpath %.iso output
vpath cfdev src/code.cloudfoundry.org/cfdev

cf-deps.iso: ./scripts/build-cf-deps-iso $(shell find src/builder ../bosh-deployment ../cf-deployment ../cf-mysql-deployment -type f)
	./scripts/build-cf-deps-iso

cfdev-efi.iso: ./scripts/build-image $(wildcard linuxkit/**/*)
	./scripts/build-image

cfdev: cf-deps.iso cfdev-efi.iso src/code.cloudfoundry.org/cfdev/generate-plugin.sh $(shell find src/code.cloudfoundry.org/{cfdev,cfdevd} -name '*.go')
	(cd src/code.cloudfoundry.org/cfdev && ./generate-plugin.sh)
