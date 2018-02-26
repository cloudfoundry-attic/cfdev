```
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

## Install 
1. Download the CF Dev binary [here](https://d3p1cc0zb2wjno.cloudfront.net/cfdev/cfdev-v0.0.1-darwin).
1. _(if needed)_ Uninstall PCF Dev plugin if its installed `cf uninstall-plugin pcfdev`.
1. Install the CF Dev plugin `cf install-plugin <path to cfdev binary>`.

## Start
Run CF Dev `cf dev start`.


## Run BOSH with CF Dev
1. _(if needed)_ Install [BOSH CLI v2](https://bosh.io/docs/cli-v2.html).
1. Set environment variables to point BOSH to your CF Dev instance `eval "$(cf dev bosh env)"`.
1. Run BOSH `bosh <command you want to run>`.

## Uninstall

To stop CF Dev run `cf dev stop`. This will completely stop and destory the CF Dev VM.

To uninstall the CF Dev cf CLI plugin run `cf uninstall-plugin cfdev`.

## Contributing

If you are interested in contributing to CF Dev, please refer to the [contributing guidelines](CONTRIBUTING.md).
