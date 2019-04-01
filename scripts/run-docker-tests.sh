#!/usr/bin/env bash

set -e

if [[ -z "$1" ]]; then
  echo "USAGE: $0 <path-to-cf-acceptance-tests-repo>"
  exit 1
fi

domain=dev.cfdev.sh
cats_path=$1

function run_docker_registry_tests() {
  export CONFIG=$(mktemp -t config.XXXXXXXX)
  export REGISTRY_AUTH_FILE=$(mktemp /tmp/registry.auth.XXXXXXXX)

  # https://docs.docker.com/registry/deploying/#native-basic-auth
  # docker run \
  #   --entrypoint htpasswd \
  #   registry:2 -Bbn testuser testpassword > auth/htpasswd

  echo 'testuser:$2y$05$AgAP8TfkxY4Fl/R7y0DOu.Ex4KGXZaHoeT2VTnFudmYoG80YGZHR.' > ${REGISTRY_AUTH_FILE}

  docker rm -f registry-cfdev || true
  docker run -d -p 5000:5000 --restart always --name registry-cfdev \
    -e REGISTRY_AUTH=htpasswd \
    -e REGISTRY_AUTH_HTPASSWD_REALM=realm \
    -e REGISTRY_AUTH_HTPASSWD_PATH=/auth/htpasswd \
    -v $REGISTRY_AUTH_FILE:/auth/htpasswd \
    registry:2

  docker pull cloudfoundry/diego-docker-app-custom:latest
  docker tag cloudfoundry/diego-docker-app-custom:latest localhost:5000/diego-docker-app-custom
  docker login localhost:5000 -u testuser -p testpassword
  docker push localhost:5000/diego-docker-app-custom

cat <<EOF >${CONFIG}
{
  "api": "api.$domain",
  "apps_domain": "$domain",
  "admin_user": "admin",
  "admin_password": "admin",
  "admin_secret": "admin-client-secret",
  "skip_ssl_validation": true,
  "use_http": true,
  "use_log_cache": false,
  "backend": "diego",
  "include_apps": false,
  "include_detect": false,
  "include_persistent_app": false,
  "include_routing": false,
  "include_v3": false,
  "include_capi_no_bridge": false,
  "include_docker": true,
  "include_private_docker_registry": true,
  "private_docker_registry_image": "host.cfdev.sh:5000/diego-docker-app-custom",
  "private_docker_registry_username": "testuser",
  "private_docker_registry_password": "testpassword",

  "default_timeout": 120,
  "cf_push_timeout": 120,
  "long_curl_timeout": 120,
  "broker_start_timeout": 120,
  "async_service_operation_timeout": 120,
  "sleep_timeout": 60
}
EOF

  pushd ${cats_path} >/dev/null
    ./bin/test -slowSpecThreshold=1200 -flakeAttempts=3 ${@:2} .
  popd >/dev/null
}


run_docker_registry_tests $@