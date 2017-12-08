FROM linuxkit/alpine:07f7d136e427dc68154cd5edbb2b9576f9ac5213 AS gaol

RUN apk add --no-cache go musl-dev git build-base
ENV GOPATH=/go PATH=$PATH:/go/bin
ENV COMMIT=616d17bec7412ec6d8adddf406036afcef3b58b6

RUN git clone https://github.com/contraband/gaol.git /go/src/github.com/contraband/gaol && \
    cd /go/src/github.com/contraband/gaol && \
    git checkout $COMMIT && \
    go install github.com/contraband/gaol

FROM linuxkit/alpine:07f7d136e427dc68154cd5edbb2b9576f9ac5213 AS build

RUN mkdir -p /out/etc/apk && cp -r /etc/apk/* /out/etc/apk/
RUN apk add --repository http://dl-3.alpinelinux.org/alpine/edge/testing/ --repository http://dl-3.alpinelinux.org/alpine/edge/main/ --repository http://dl-3.alpinelinux.org/alpine/edge/community/ --allow-untrusted --no-cache --initdb -p /out \
    alpine-baselayout \
    bash \
    btrfs-progs \
    busybox \
    curl \
    file \
    findutils \
    iptables \
    util-linux

RUN apk --no-cache add ca-certificates curl && \
  curl -L -o /tmp/glibc.apk https://github.com/sgerrand/alpine-pkg-glibc/releases/download/2.26-r0/glibc-2.26-r0.apk && \
  apk add --allow-untrusted -p /out /tmp/glibc.apk

RUN rm -rf /out/etc/apk /out/lib/apk /out/var/cache

FROM scratch
WORKDIR /
COPY --from=build /out/ /
COPY --from=gaol  /go/bin/gaol /usr/bin/gaol

ENV GROOTFS_VERSION 0.28.1
ENV DRAX_SHA256     cae527d5afd19139f2b29f742fcd47c4fe2d5b805784de337904b6abeb1e450e
ENV TARDIS_SHA256   97184e8db8348d89b433f9fcf63877344ef3664206ff6803f78eb78b6a8ac998
ENV GROOTFS_SHA256  4c85123553faa6e34184f6c7771a5ada4b5d1a09f560b8312db21dd794ec631f

ENV GDN_VERSION 1.10.0
ENV GDN_SHA256  303717d871c95613f351d54a2d8ff71c381f613c185d0541b0009eacc5dbccf9

RUN set -x \
    && curl -fSL "https://github.com/cloudfoundry/grootfs/releases/download/v${GROOTFS_VERSION}/drax-${GROOTFS_VERSION}" -o drax \
    && echo "${DRAX_SHA256} *drax" | sha256sum -c - \
    && mv drax /usr/bin/drax \
    && chmod a+x /usr/bin/drax \
    && curl -fSL "https://github.com/cloudfoundry/grootfs/releases/download/v${GROOTFS_VERSION}/tardis-${GROOTFS_VERSION}" -o tardis \
    && echo "${TARDIS_SHA256} *tardis" | sha256sum -c - \
    && mv tardis /usr/bin/tardis \
    && chmod a+x /usr/bin/tardis \
    && curl -fSL "https://github.com/cloudfoundry/grootfs/releases/download/v${GROOTFS_VERSION}/grootfs-${GROOTFS_VERSION}" -o grootfs \
    && echo "${GROOTFS_SHA256} *grootfs" | sha256sum -c - \
    && mv grootfs /usr/bin/grootfs \
    && chmod a+x /usr/bin/grootfs \
    && curl -fSL "https://github.com/cloudfoundry/garden-runc-release/releases/download/v${GDN_VERSION}/gdn-${GDN_VERSION}" -o gdn \
    && echo "${GDN_SHA256} *gdn" | sha256sum -c - \
    && mv gdn /usr/bin/gdn \
    && chmod a+x /usr/bin/gdn

COPY rootfs /

ENTRYPOINT ["/usr/bin/garden"]
CMD []
LABEL org.mobyproject.config='{ \
    "net": "host", \
    "capabilities" : ["all"], \
    "binds": ["/etc/resolv.conf:/etc/resolv.conf", "/lib/modules:/lib/modules"], \
    "mounts": [{"type":"cgroup", "options":["rw", "nosuid","noexec","nodev","relatime"]}] \
}'
