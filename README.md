```
$ cf dev start
Downloading Resources...
Starting VPNKit ...
Starting the VM...
Deploying the BOSH Director...
Deploying CF...

  ██████╗███████╗██████╗ ███████╗██╗   ██╗
 ██╔════╝██╔════╝██╔══██╗██╔════╝██║   ██║
 ██║     █████╗  ██║  ██║█████╗  ██║   ██║
 ██║     ██╔══╝  ██║  ██║██╔══╝  ╚██╗ ██╔╝
 ╚██████╗██║     ██████╔╝███████╗ ╚████╔╝
  ╚═════╝╚═╝     ╚═════╝ ╚══════╝  ╚═══╝
             is now running!

To begin using CF Dev, please run:
    cf login -a https://api.v3.pcfdev.io --skip-ssl-validation

Admin user => Email: admin / Password: admin
Regular user => Email: user / Password: pass

```

***********************************

CF Dev is a new distribution of Cloud Foundry designed to run on a developer’s laptop or workstation using native hypervisors and a fully functional BOSH Director. CF Dev gives application developers the full Cloud Foundry experience in a lightweight, easy to install package. CF Dev is intended for application developers who wish to develop and debug their application locally on a full-featured Cloud Foundry. CF Dev is also an excellent getting started environment for developers interested in learning and exploring Cloud Foundry.

## Prerequisites

* [CF CLI](https://github.com/cloudfoundry/cli)
* Internet connection (or Dnsmasq or Acrylic) required for wildcard DNS resolution
* Please note CF Dev only supports macOS at this time

## Install 
1. _(if needed)_ Uninstall your existing PCF Dev plugin if it is installed `cf uninstall-plugin pcfdev`
1. Install the CF Dev plugin `cf install-plugin cfdev`.

## Start
Run CF Dev `cf dev start`.


## Run BOSH with CF Dev
1. _(if needed)_ Install [BOSH CLI v2](https://bosh.io/docs/cli-v2.html).
1. Set environment variables to point BOSH to your CF Dev instance `eval "$(cf dev bosh env)"`.
1. Run BOSH `bosh <command you want to run>`.

## Project Backlog

Follow the CF Dev team's progress [here](https://github.com/cloudfoundry-incubator/cfdev/projects/1).  This backlog contains a prioritized list of features and bugs the CF Dev team is working on.  Check the project board for the latest updates on features and when they will be released.

## Uninstall

To stop CF Dev run `cf dev stop`. This will completely stop and destroy the CF Dev VM.

To uninstall the CF Dev cf CLI plugin run `cf uninstall-plugin cfdev`.

## Telemetry

Here on the CF Dev team, we use telemetry to help us understand how our tool is being used.  We value our users privacy, therefore all telemetry is completely anonymous. There is no way for anyone with the telemetry to identify who is using the CF Dev tool.  In an effort to make our data as transparent as possible, we will be publishing aggregated anonymous usage data to this page periodically to help our user community understand how the tool is being used. 

In addition to making this data completely anonymous, we require users to opt-in to allowing us to collect telemetry from their tool. Upon running `$ cf dev start` for the first time, we will prompt the user to opt-in to capturing analytics.  Any time after that you can turn on/off telemetry by running `$ cf dev telemetry --on/off`

You can learn more about what we do with telemetry [here](https://github.com/cloudfoundry-incubator/cfdev/wiki/Telemetry)

## TCP Ports

The tcp port range has been limited to 1024 - 1049 to prevent reaching file descriptor limits on some systems.

## Contributing

If you are interested in contributing to CF Dev, please refer to the [contributing guidelines](CONTRIBUTING.md).
