#!/bin/bash

gaol create -n deploy-bosh -p \
  --network 10.246.0.0/16 \
  -r /var/vcap/director/cache/deploy-bosh.tar \
  -m /var/vcap:/var/vcap \
  -m /var/vcap/director/cache:/var/vcap/director/cache

gaol run deploy-bosh --attach -c /usr/bin/deploy-bosh
