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

COPY --from=bosh-cli-builder /go/src/github.com/cloudfoundry/bosh-cli/out/bosh /usr/bin/bosh

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

ENV BOSH_DEPLOYMENT_COMMIT 9d3cd699f6eb905cb763df967e8b7e6f16510d27

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

ADD operations/* /etc/bosh/deployment/
ADD deploy-bosh /usr/bin/deploy-bosh

RUN cd /etc/bosh/deployment && \
    bosh int bosh.yml \
      -o bosh-lite.yml \
      -o bosh-lite-runc.yml \
      -o bosh-lite-grootfs.yml \
      -o warden/cpi.yml \
      -o warden/cpi-grootfs.yml \
      -o jumpbox-user.yml \
      -o disable-app-armor.yml \
      -o remove-ports.yml \
      -o use-warden-cpi-v39.yml \
      -o use-garden-1_10.yml \
      -v director_name="warden" \
      -v internal_cidr=10.245.0.0/24 \
      -v internal_gw=10.245.0.1 \
      -v internal_ip=10.245.0.2 \
      -v garden_host=10.0.0.10 \
      > "/etc/bosh/director.yml" && \
    \
    echo -e "\n# bosh-deployment commit: ${BOSH_DEPLOYMENT_COMMIT}" \
      >> /etc/bosh/director.yml

RUN mkdir -p /var/vcap

CMD ["/usr/bin/deploy-bosh"]
