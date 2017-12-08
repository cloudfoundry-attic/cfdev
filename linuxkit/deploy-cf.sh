#!/bin/bash

gaol create -n deploy-cf -p \
  --network 10.246.0.0/16 \
  -r /var/vcap/cf/cache/deploy-cf.tar \
  -m /var/vcap:/var/vcap \
  -m /var/vcap/cf/cache:/var/vcap/cf/cache

gaol run deploy-cf --attach -c /usr/bin/deploy-cf
