FROM linuxkit/alpine:07f7d136e427dc68154cd5edbb2b9576f9ac5213 AS build

RUN mkdir -p /out/etc/apk && cp -r /etc/apk/* /out/etc/apk/

RUN apk add \
  --repository http://dl-3.alpinelinux.org/alpine/edge/testing/ \
  --repository http://dl-3.alpinelinux.org/alpine/edge/main/ \
  --repository http://dl-3.alpinelinux.org/alpine/edge/community/ \
  \
  --allow-untrusted --no-cache --initdb -p /out \
  \
  busybox \
  inotify-tools \
  iptables

RUN rm -rf /out/etc/apk /out/lib/apk /out/var/cache

FROM scratch
WORKDIR /
COPY --from=build /out/ /
COPY setup-bosh-lite-routing /usr/bin/setup-bosh-lite-routing
CMD ["/usr/bin/setup-bosh-lite-routing"]