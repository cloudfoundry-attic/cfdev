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

## Local scripts & dependencies
ADD app-security-group.json /etc/cf/app-security-group.json
ADD deploy-bosh /usr/bin/deploy-bosh
ADD deploy-cf /usr/bin/deploy-cf

RUN mkdir -p /var/vcap

# TODO Throw an error if BOSH & CF deployment do not have the same major version
