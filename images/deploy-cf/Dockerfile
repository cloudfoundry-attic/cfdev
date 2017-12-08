FROM ruby:2.4.2-alpine3.6

RUN set -x && \
    apk add --no-cache \
      bash \
      ca-certificates \
      coreutils \
      curl \
      openssh && \
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

ENV CF_CLI_VERSION 6.32.0
ENV CF_CLI_SHA256 de3f7cf48cefac8f42b908c4e32f01b37b90a6f0d8ff7353ac884a7b00ff0e08

RUN set -x && \
    curl -fsSL "https://packages.cloudfoundry.org/stable?release=linux64-binary&version=${CF_CLI_VERSION}&source=github-rel" \
        | tar -xz -C /usr/bin cf && \
    echo "${CF_CLI_SHA256} /usr/bin/cf" | sha256sum -c - && \
    \
    chmod a+x /usr/bin/cf

ENV CF_DEPLOYMENT_COMMIT 9be2644da8de08540891e24856bbdb88f9a83f67

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


ADD operations/* /etc/cf/deployment/
ADD app-security-group.json /etc/cf/app-security-group.json
ADD deploy-cf /usr/bin/deploy-cf

#-o enable-volume-service-local.yml
#-o allow-local-docker-registry.yml

RUN cd /etc/cf/deployment && \
      bosh int cf-deployment.yml \
      -o operations/use-compiled-releases.yml \
      -o operations/enable-privileged-container-support.yml \
      -o operations/experimental/use-grootfs.yml \
      -o operations/bosh-lite.yml \
      -o garden-disable-app-armour.yml \
      -o collocate-tcp-router.yml \
      -o use-btrfs-grootfs.yml \
      -v cf_admin_password=admin \
      -v uaa_admin_client_secret=admin-client-secret \
      > "/etc/cf/deployment.yml" && \
    \
    echo -e "\n# cf-deployment commit: ${CF_DEPLOYMENT_COMMIT}" \
      >> /etc/cf/deployment.yml

RUN mkdir -p /var/vcap

CMD ["/usr/bin/deploy-cf"]
