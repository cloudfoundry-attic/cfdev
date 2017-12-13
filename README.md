

### Build & Test Dependencies
- Docker for Mac
- Linuxkit - https://github.com/linuxkit/linuxkit
- brew install cdrtools
- garden cli - https://github.com/contraband/gaol

### Running the VM manually

In linuxkit folder run the following:
- build-image.sh - builds the vm iso
- build-bosh-iso.sh
- build-cf-iso.sh
- build-firmware.sh
- fetch-executables.sh
- setup-ip-aliases.sh
- run.sh
- deploy-bosh.sh
- deploy-cf.sh

### Running tests

- Remove IP aliases prior to running tests. Use `linuxkit/teardown-ip-aliases.sh`
- See `src/code.cloudfoundry.org/cfdev/run-tests.sh` to see which tests need require sudo (root access)
