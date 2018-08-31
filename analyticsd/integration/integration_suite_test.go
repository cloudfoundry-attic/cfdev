package integration

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestIntegration(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Integration Suite")
}

var fixturePushApp = `
{
  "total_results": 1,
  "total_pages": 1,
  "prev_url": null,
  "next_url": null,
  "resources": [
    {
      "metadata": {
        "guid": "84e40957-9c33-4662-a123-f3c89f0dc254",
        "url": "/v2/events/84e40957-9c33-4662-a123-f3c89f0dc254",
        "created_at": "2018-08-20T20:20:43Z",
        "updated_at": "2018-08-20T20:20:43Z"
      },
      "entity": {
        "type": "audit.app.create",
        "actor": "949d40cf-ff27-43a9-b7f3-15df6573d165",
        "actor_type": "user",
        "actor_name": "admin",
        "actor_username": "admin",
        "actee": "4ecf7116-b7aa-4b3d-abfd-e69c5f267a32",
        "actee_type": "app",
        "actee_name": "dora",
        "timestamp": "2018-08-08T08:08:08Z",
        "metadata": {
          "request": {
            "buildpack": "ruby_buildpack",
            "name": "dora",
            "space_guid": "6f47577d-0a45-4415-91a2-7456c4811398",
            "console": false,
            "docker_credentials": "[PRIVATE DATA HIDDEN]",
            "environment_json": "[PRIVATE DATA HIDDEN]",
            "health_check_type": "port",
            "instances": 1,
            "production": false,
            "state": "STOPPED"
          }
        },
        "space_guid": "6f47577d-0a45-4415-91a2-7456c4811398",
        "organization_guid": "4f05fbda-7fa4-4fb1-b705-5826cbe6ef2f"
      }
    },
    {
      "metadata": {
        "guid": "84e40957-9c33-4662-a123-f3c89f0dc254",
        "url": "/v2/events/84e40957-9c33-4662-a123-f3c89f0dc254",
        "created_at": "2018-08-20T20:20:43Z",
        "updated_at": "2018-08-20T20:20:43Z"
      },
      "entity": {
        "type": "audit.app.create",
        "actor": "949d40cf-ff27-43a9-b7f3-15df6573d165",
        "actor_type": "user",
        "actor_name": "admin",
        "actor_username": "admin",
        "actee": "4ecf7116-b7aa-4b3d-abfd-e69c5f267a32",
        "actee_type": "app",
        "actee_name": "dora2",
        "timestamp": "2018-09-09T09:09:09Z",
        "metadata": {
          "request": {
            "buildpack": "go_buildpack",
            "name": "dora",
            "space_guid": "6f47577d-0a45-4415-91a2-7456c4811398",
            "console": false,
            "docker_credentials": "[PRIVATE DATA HIDDEN]",
            "environment_json": "[PRIVATE DATA HIDDEN]",
            "health_check_type": "port",
            "instances": 1,
            "production": false,
            "state": "STOPPED"
          }
        },
        "space_guid": "6f47577d-0a45-4415-91a2-7456c4811398",
        "organization_guid": "4f05fbda-7fa4-4fb1-b705-5826cbe6ef2f"
      }
    }
  ]
}
`

var fixtureSequentialResponse1 = `
{
  "total_results": 1,
  "total_pages": 1,
  "prev_url": null,
  "next_url": null,
  "resources": [
    {
      "metadata": {
        "guid": "84e40957-9c33-4662-a123-f3c89f0dc254",
        "url": "/v2/events/84e40957-9c33-4662-a123-f3c89f0dc254",
        "created_at": "2018-08-20T20:20:43Z",
        "updated_at": "2018-08-20T20:20:43Z"
      },
      "entity": {
        "type": "audit.app.create",
        "actor": "949d40cf-ff27-43a9-b7f3-15df6573d165",
        "actor_type": "user",
        "actor_name": "admin",
        "actor_username": "admin",
        "actee": "4ecf7116-b7aa-4b3d-abfd-e69c5f267a32",
        "actee_type": "app",
        "actee_name": "dora",
        "timestamp": "2018-08-08T08:08:08Z",
        "metadata": {
          "request": {
            "buildpack": "ruby_buildpack",
            "name": "dora",
            "space_guid": "6f47577d-0a45-4415-91a2-7456c4811398",
            "console": false,
            "docker_credentials": "[PRIVATE DATA HIDDEN]",
            "environment_json": "[PRIVATE DATA HIDDEN]",
            "health_check_type": "port",
            "instances": 1,
            "production": false,
            "state": "STOPPED"
          }
        },
        "space_guid": "6f47577d-0a45-4415-91a2-7456c4811398",
        "organization_guid": "4f05fbda-7fa4-4fb1-b705-5826cbe6ef2f"
      }
    }
  ]
}
`

var fixtureSequentialResponse2 = `
{
  "total_results": 2,
  "total_pages": 1,
  "prev_url": null,
  "next_url": null,
  "resources": [
    {
      "metadata": {
        "guid": "84e40957-9c33-4662-a123-f3c89f0dc254",
        "url": "/v2/events/84e40957-9c33-4662-a123-f3c89f0dc254",
        "created_at": "2018-08-20T20:20:43Z",
        "updated_at": "2018-08-20T20:20:43Z"
      },
      "entity": {
        "type": "audit.app.create",
        "actor": "949d40cf-ff27-43a9-b7f3-15df6573d165",
        "actor_type": "user",
        "actor_name": "admin",
        "actor_username": "admin",
        "actee": "4ecf7116-b7aa-4b3d-abfd-e69c5f267a32",
        "actee_type": "app",
        "actee_name": "dora",
        "timestamp": "2018-08-08T08:08:09Z",
        "metadata": {
          "request": {
            "buildpack": "go_buildpack",
            "name": "dora",
            "space_guid": "6f47577d-0a45-4415-91a2-7456c4811398",
            "console": false,
            "docker_credentials": "[PRIVATE DATA HIDDEN]",
            "environment_json": "[PRIVATE DATA HIDDEN]",
            "health_check_type": "port",
            "instances": 1,
            "production": false,
            "state": "STOPPED"
          }
        },
        "space_guid": "6f47577d-0a45-4415-91a2-7456c4811398",
        "organization_guid": "4f05fbda-7fa4-4fb1-b705-5826cbe6ef2f"
      }
    }
  ]
}
`

var fixtureUnexpected = `
{
  "total_results": 1,
  "total_pages": 1,
  "prev_url": null,
  "next_url": null,
  "resources": [
    {
      "metadata": {},
      "entity": {
        "type": "some-unexpected-event-type"
      }
    }
  ]
}
`