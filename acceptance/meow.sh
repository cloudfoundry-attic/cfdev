#!/bin/bash

set -e

script_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
domain=${1-v3.pcfdev.io}

function run_cats() {
  export CONFIG=$(mktemp -t config.XXXXXXXX)
  cat <<EOF >${CONFIG}
  {
    "api": "api.$domain",
    "apps_domain": "$domain",
    "admin_user": "admin",
    "admin_password": "admin",
    "admin_secret": "admin-client-secret",
    "skip_ssl_validation": true,
    "use_http": true,
    "backend": "diego",
    "include_apps": true,
    "include_backend_compatibility": false,
    "include_container_networking": true,
    "include_capi_no_bridge" : true,
    "include_detect": true,
    "include_docker": true,
    "include_internet_dependent": true,
    "include_persistent_app": true,
    "include_privileged_container_support": true,
    "include_route_services": true,
    "include_routing": true,
    "include_security_groups": true,
    "include_services": true,
    "include_ssh": true,
    "include_sso": true,
    "include_tasks": true,
    "include_v3": true,
    "default_timeout": 120,
    "cf_push_timeout": 120,
    "long_curl_timeout": 120,
    "broker_start_timeout": 120,
    "async_service_operation_timeout": 120,
    "sleep_timeout": 60
  }
EOF

  pushd $script_dir/src/github.com/cloudfoundry/cf-acceptance-tests >/dev/null
    GOPATH=$script_dir ./bin/test -slowSpecThreshold=120 ${@:2} .
  popd >/dev/null

}

function run_networking_tests() {
  networking_release_path="$script_dir/src/github.com/cloudfoundry/cf-networking-release"

  export CONFIG=$(mktemp -t config.XXXXXXXX)
  export APPS_DIR="$networking_release_path/src/example-apps"

  pushd "$networking_release_path/src/test/acceptance" > /dev/null

cat <<EOF >${CONFIG}
{
  "api": "api.${domain}",
  "admin_user": "admin",
  "admin_password": "admin",
  "admin_secret": "admin-client-secret",
  "apps_domain": "${domain}",
  "default_security_groups": [ "dns", "public_networks" ],
  "skip_ssl_validation": true,
  "test_app_instances": 2,
  "test_applications": 2,
  "proxy_instances": 1,
  "proxy_applications": 1,
  "extra_listen_ports": 2,
  "prefix":"cf-networking"
}
EOF

    export GOPATH=$networking_release_path
    export PATH=$networking_release_path/bin:$PATH
    ginkgo -slowSpecThreshold=120 ${@:2} .
  popd > /dev/null
}

function run_routing_tests() {
  export CONFIG=$(mktemp -t config.XXXXXXXX)

  gaol create -p -n fetch-cf-vars \
    -r /var/vcap/director/cache/deploy-bosh.tar \
    -m /var/vcap:/var/vcap

  local gaol_output=$(gaol run fetch-cf-vars \
    --attach \
    --command "grep uaa_clients_tcp_emitter_secret /var/vcap/cf/vars.yml")

  local tcp_emitter_password=$(echo $gaol_output | cut -d' ' -f 2)

  gaol destroy fetch-cf-vars

cat <<EOF >$CONFIG
{
  "addresses": ["$domain"],
  "admin_user": "admin",
  "admin_password": "admin",
  "api": "api.$domain",
  "apps_domain": "$domain",
  "cf_push_timeout": 480,
  "default_timeout": 480,
  "include_http_routes": true,
  "skip_ssl_validation": true,
  "tcp_router_group":"default-tcp",
  "test_password": "test",
  "use_http":true,
  "oauth": {
    "token_endpoint": "https://uaa.$domain",
    "client_name": "tcp_emitter",
    "client_secret": "$tcp_emitter_password",
    "port": 443,
    "skip_ssl_validation": true
  }
}
EOF

  local release_path="$script_dir/src/github.com/cloudfoundry-incubator/routing-release"
  local test_path="$release_path/src/code.cloudfoundry.org/routing-acceptance-tests"

  pushd "$test_path" >/dev/null
    export GOPATH=$release_path
    export PATH=$GOPATH/bin:$PATH

    pushd $GOPATH/src/code.cloudfoundry.org/routing-api-cli > /dev/null
      go build -ldflags "-X main.version=2.9" -o $GOPATH/bin/rtr
    popd > /dev/null

    ginkgo -r -race -slowSpecThreshold=120 ${@:2} smoke_tests http_routes tcp_routing
  popd >/dev/null
}

function run_persi_tests() {
  volume_release_path="$script_dir/src/github.com/cloudfoundry/local-volume-release"
  persi_test_dir="$volume_release_path/src/code.cloudfoundry.org/persi-acceptance-tests"

  export CONFIG=$(mktemp -t config.XXXXXXXX)

cat <<EOF >$CONFIG
{
  "admin_user": "admin",
  "admin_password": "admin",
  "api": "api.$domain",
  "apps_domain": "$domain",
  "prefix": "persi-",
  "skip_ssl_validation": true,
  "default_timeout": 30,
  "broker_url": "http://local-broker.$domain",
  "broker_user": "username",
  "broker_password": "password",
  "service_name": "local-volume",
  "plan_name": "free"
}
EOF

  pushd "$persi_test_dir" >/dev/null
    export GOPATH=$volume_release_path
    export TEST_APPLICATION_PATH="$persi_test_dir/assets/pora"

    ginkgo -r -slowSpecThreshold=120 ${@:2} .
  popd >/dev/null
}


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
  "backend": "diego",
  "include_apps": false,
  "include_detect": false,
  "include_persistent_app": false,
  "include_routing": false,
  "include_v3": false,
  "include_capi_no_bridge": false,
  "include_docker": true,
  "include_private_docker_registry": true,
  "private_docker_registry_image": "host.pcfdev.io:5000/diego-docker-app-custom",
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


  pushd src/github.com/cloudfoundry/cf-acceptance-tests >/dev/null
    # Apply our patches
    # Remove this patch when cf-deployment's corresponding cats has
    # https://github.com/cloudfoundry/cf-acceptance-tests/pull/262
    git checkout .
    git apply $script_dir/patches/private-docker-registry-auth.patch

    GOPATH=$script_dir ./bin/test -slowSpecThreshold=120 ${@:2} .

    # Undo our patches
    git apply -R $script_dir/patches/private-docker-registry-auth.patch

  popd >/dev/null
}

# Remove PCF Dev 'all_access' application security group

export CF_HOME=$(mktemp -d)
cf api api.$domain --skip-ssl-validation
cf auth admin admin

cf unbind-staging-security-group all_access
cf unbind-running-security-group all_access

trap "{ \
  cf bind-staging-security-group all_access; \
  cf bind-running-security-group all_access; \
}" EXIT


run_cats $@
run_networking_tests $@
run_routing_tests $@
#run_persi_tests $@

# Docker registry will run on a local IP so we
# allow containers to access internal networks again
cf bind-staging-security-group all_access
cf bind-running-security-group all_access

run_docker_registry_tests $@

# not enabled for cats
# include_credhub
# include_capi_experimental
# include_service_instance_sharing
# include_zipkin
# include_isolation_segments
# include_routing_isolation_segments