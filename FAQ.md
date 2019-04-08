# FAQ

## Can I configure the working directory?

Yes you can, by exporting the environment variable `CFDEV_HOME` to your custom location before invoking CF Dev. All 
state used by the plugin will be contained in the that directory. By default, the working directory is set to `~/.cfdev`.

## `cf dev start` failed while deploying. It got stuck at _Progress x of 15 (55m10s)_.

Running an entire PAAS on one workstation is an ambitious endeavor and the workstation has to be performant enough
for the task. The most common insufficiency is the disk speed due to CF Dev being such an disk I/O intensive process.
As mentioned in our [recommended requirements](https://github.com/cloudfoundry-incubator/cfdev#recommended-system-requirements).
No less than flash storage (i.e. SSDs) is recommended for use with CF Dev. If your workstation has an HDD, this is the most
likely cause of issue.

We are always working towards reducing the footprint of CF Dev.

## The only service available is mysql. How do I get access to pivotal apps manager, rabbitmq, redis, spring-cloud-services?

A separate asset is needed. You can download the correct asset for your platform at
[https://network.pivotal.io/products/pcfdev](https://network.pivotal.io/products/pcfdev).
Then you perform a start with the downloaded asset specified via the `-f` flag, like so: `cf dev start -f ./pcfdev-v*.tgz`

## Why are the assets to download so big?

Under the hood, CF Dev is performing a [BOSH](https://bosh.io/docs/) deploy - but into containers rather than full-sized VMs.
The significant bulk of what is packaged are those same assets that are needed for a deployment into a cloud provider. 

## How do I inspect debugging information?

Logging information is written to various files in the `log` directory of the CF Dev working directory. If the working directory has not
been reconfigured, it can be found in: `~/.cfdev/log`.

## What ports are available for use with TCP routing?

`1024 - 1049`.

## Is there a command to make more ports available?

No. Exposing ports to the VM can currently only be specified during the build process of the virtual machine.

## What is the pcfdev repository? https://github.com/pivotal-cf/pcfdev

pcfdev is our previous, _deprecated_ offering that satisfied the same use case: standing up a local Cloud Foundry environment.
We have chosen to move away from it because its architecture made maintenance incredibly difficult. Engineers are no longer allocated to it
and CF Dev is meant to supplant it completely. In addition to offering more recent version of CF internals and features, CF Dev
make use of native hypervisors which offers better performance.

## Copyright

See [LICENSE](LICENSE) for details.
Copyright (c) 2018 [Pivotal Software, Inc](http://www.pivotal.io/).
