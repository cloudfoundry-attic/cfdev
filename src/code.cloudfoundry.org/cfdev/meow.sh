#!/bin/bash

set -e

if [[ -z "$1" ]]; then
  echo "USAGE: $0 <path-to-cf-acceptance-tests-repo>"
  exit 1
fi

domain=dev.cfdev.sh
cats_path=$1

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
    "include_privileged_container_support": false,
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

  pushd ${cats_path} >/dev/null
    ./bin/test -slowSpecThreshold=1200 --flakeAttempts=3 ${@:2} .
  popd >/dev/null

}

# Remove PCF Dev 'all_access' application security group
export CF_HOME=$(mktemp -d)
cf api api.$domain --skip-ssl-validation
cf auth admin admin

cf unbind-staging-security-group all_access
cf unbind-running-security-group all_access

run_cats $@

# Re-enable PCF Dev 'all_access' application security group
export CF_HOME=$(mktemp -d)
cf api api.$domain --skip-ssl-validation
cf auth admin admin

cf bind-staging-security-group all_access
cf bind-running-security-group all_access
