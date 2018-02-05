FROM golang:1.9-alpine3.6 as bosh-cli-builder

ENV BOSH_CLI_COMMIT 15dd4c843e161ed6410b4d91947484b3b87f39ce

RUN set -x && \
    apk add --no-cache curl bash && \
    mkdir -p /go/src/github.com/cloudfoundry && \
    \
    curl -fsSL https://github.com/pcfdev-forks/bosh-cli/archive/${BOSH_CLI_COMMIT}.zip \
       -o /cli.zip && \
    \
    cd /go/src/github.com/cloudfoundry && \
    \
    unzip /cli.zip -d $PWD && \
    \
    rm /cli.zip && \
    \
    mv bosh-cli-${BOSH_CLI_COMMIT} bosh-cli && \
    \
    chmod +x bosh-cli/bin/build && \
    \
    bosh-cli/bin/build

FROM ruby:2.4.2-alpine3.6

RUN set -x && \
    apk add --no-cache \
      bash \
      ca-certificates \
      coreutils \
      curl \
      e2fsprogs \
      iptables \
      openssh \
      sudo \
      util-linux && \
    \
    curl -fsSL https://github.com/sgerrand/alpine-pkg-glibc/releases/download/2.26-r0/glibc-2.26-r0.apk \
       -o /tmp/glibc.apk && \
    \
    apk add --no-cache --allow-untrusted /tmp/glibc.apk && \
    rm /tmp/glibc.apk

### CLIs
COPY --from=bosh-cli-builder /go/src/github.com/cloudfoundry/bosh-cli/out/bosh /usr/bin/bosh

ENV CF_CLI_VERSION 6.32.0
ENV CF_CLI_SHA256 de3f7cf48cefac8f42b908c4e32f01b37b90a6f0d8ff7353ac884a7b00ff0e08

RUN set -x && \
    curl -fsSL "https://packages.cloudfoundry.org/stable?release=linux64-binary&version=${CF_CLI_VERSION}&source=github-rel" \
        | tar -xz -C /usr/bin cf && \
    echo "${CF_CLI_SHA256} /usr/bin/cf" | sha256sum -c - && \
    \
    chmod a+x /usr/bin/cf


## Deployments
ENV BOSH_DEPLOYMENT_COMMIT 218e6d5030d89ca9f31c50b8b308e2a78d2a0997

RUN set -x && \
    mkdir -p /etc/bosh && \
    \
    curl -fsSL https://github.com/cloudfoundry/bosh-deployment/archive/${BOSH_DEPLOYMENT_COMMIT}.zip \
       -o /etc/bosh/deployment.zip && \
    \
    unzip /etc/bosh/deployment.zip -d /etc/bosh && \
    \
    rm /etc/bosh/deployment.zip && \
    \
    mv /etc/bosh/bosh-deployment-${BOSH_DEPLOYMENT_COMMIT} /etc/bosh/deployment && \
    \
    echo "${BOSH_DEPLOYMENT_COMMIT}" > /etc/bosh/deployment/commit

ADD bosh-operations /etc/bosh/deployment/

RUN cd /etc/bosh/deployment && \
    bosh int bosh.yml \
      -o bosh-lite.yml \
      -o bosh-lite-runc.yml \
      -o bosh-lite-grootfs.yml \
      -o warden/cpi.yml \
      -o warden/cpi-grootfs.yml \
      \
      -o jumpbox-user.yml \
      -o disable-app-armor.yml \
      -o remove-ports.yml \
      -o use-warden-cpi-v39.yml \
      -o use-stemcell-3468.17.yml \
      \
      -v director_name="warden" \
      -v internal_cidr=10.245.0.0/24 \
      -v internal_gw=10.245.0.1 \
      -v internal_ip=10.245.0.2 \
      -v garden_host=10.0.0.10 \
      > "/etc/bosh/director.yml" && \
    \
    echo -e "\n# bosh-deployment commit: ${BOSH_DEPLOYMENT_COMMIT}" \
      >> /etc/bosh/director.yml


ENV CF_DEPLOYMENT_COMMIT e25735313e7fdc4508222c760e93f38f2e9c4812

RUN set -x && \
    mkdir -p /etc/cf && \
    \
    curl -fsSL https://github.com/cloudfoundry/cf-deployment/archive/${CF_DEPLOYMENT_COMMIT}.zip \
       -o /etc/cf/deployment.zip && \
    \
    unzip /etc/cf/deployment.zip -d /etc/cf && \
    \
    rm /etc/cf/deployment.zip && \
    \
    mv /etc/cf/cf-deployment-${CF_DEPLOYMENT_COMMIT} /etc/cf/deployment && \
    \
    echo "${CF_DEPLOYMENT_COMMIT}" > /etc/cf/deployment/commit

ADD cf-operations /etc/cf/deployment/

#-o enable-volume-service-local.yml
RUN cd /etc/cf/deployment && \
    bosh int iaas-support/bosh-lite/cloud-config.yml \
      -o set-cloud-config-subnet.yml \
      > "/etc/cf/cloud-config.yml" && \
    \
    echo -e "\n# cf-deployment commit: ${CF_DEPLOYMENT_COMMIT}" \
      >> /etc/cf/cloud-config.yml

RUN cd /etc/cf/deployment && \
      bosh int cf-deployment.yml \
      -o operations/use-compiled-releases.yml \
      -o operations/enable-privileged-container-support.yml \
      -o operations/experimental/use-grootfs.yml \
      \
      -o operations/experimental/skip-consul-cell-registrations.yml \
      -o operations/experimental/skip-consul-locks.yml \
      -o operations/experimental/use-bosh-dns.yml \
      -o operations/experimental/use-bosh-dns-for-containers.yml \
      -o operations/experimental/disable-consul.yml \
      -o operations/bosh-lite.yml \
      -o operations/experimental/disable-consul-bosh-lite.yml \
      \
      -o allow-local-docker-registry.yml \
      -o add-host-pcfdev-dns-record.yml \
      -o garden-disable-app-armour.yml \
      -o collocate-tcp-router.yml \
      -o use-btrfs-grootfs.yml \
      -o set-cfdev-subnet.yml \
      \
      -v cf_admin_password=admin \
      -v uaa_admin_client_secret=admin-client-secret \
      > "/etc/cf/deployment.yml" && \
    \
    echo -e "\n# cf-deployment commit: ${CF_DEPLOYMENT_COMMIT}" \
      >> /etc/cf/deployment.yml

## Local scripts & dependencies
ADD app-security-group.json /etc/cf/app-security-group.json
ADD deploy-bosh /usr/bin/deploy-bosh
ADD deploy-cf /usr/bin/deploy-cf

RUN mkdir -p /var/vcap

# TODO Throw an error if BOSH & CF deployment do not have the same major version
