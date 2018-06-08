FROM golang:1.9-alpine3.6 as bosh-cli-builder

ENV BOSH_CLI_COMMIT 052d4add5d2d5d3501b4ace855f41b826bbdc2e6

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
      jq \
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
ENV CF_CLI_VERSION 6.32.0
ENV CF_CLI_SHA256 de3f7cf48cefac8f42b908c4e32f01b37b90a6f0d8ff7353ac884a7b00ff0e08
COPY --from=bosh-cli-builder /go/src/github.com/cloudfoundry/bosh-cli/out/bosh /usr/bin/bosh

RUN set -x && \
    curl -fsSL "https://packages.cloudfoundry.org/stable?release=linux64-binary&version=$CF_CLI_VERSION&source=github-rel" \
        | tar -xz -C /usr/bin cf && \
    echo "$CF_CLI_SHA256 /usr/bin/cf" | sha256sum -c - && \
    \
    chmod a+x /usr/bin/cf

RUN mkdir -p /var/vcap

# TODO Throw an error if BOSH & CF deployment do not have the same major version
