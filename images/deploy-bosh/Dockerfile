FROM ruby:2.4.2-alpine3.6

RUN set -x && \
    apk add --no-cache \
      bash \
      ca-certificates \
      coreutils \
      curl \
      e2fsprogs \
      iptables \
      sudo \
      util-linux && \
    \
    curl -fsSL https://github.com/sgerrand/alpine-pkg-glibc/releases/download/2.26-r0/glibc-2.26-r0.apk \
       -o /tmp/glibc.apk && \
    \
    apk add --no-cache --allow-untrusted /tmp/glibc.apk && \
    rm /tmp/glibc.apk

ENV BOSH_CLI_VERSION 2.0.44
ENV BOSH_CLI_SHA256 97add408895a5d47d5524fe3eb67cb492dff7c259fe4b8afe29bdefbc88532d9

RUN set -x && \
    curl -fsSL "https://s3.amazonaws.com/bosh-cli-artifacts/bosh-cli-${BOSH_CLI_VERSION}-linux-amd64" \
       -o /usr/bin/bosh && \
    echo "${BOSH_CLI_SHA256} /usr/bin/bosh" | sha256sum -c - && \
    \
    chmod a+x /usr/bin/bosh

ENV BOSH_DEPLOYMENT_COMMIT 870a6744615120236ca78239da3a64dfa95d87b6

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
      -o remote-ports.yml \
      -o use-warden-cpi-v39.yml \
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
