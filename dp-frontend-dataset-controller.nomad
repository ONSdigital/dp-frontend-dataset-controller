job "dp-frontend-dataset-controller" {
  datacenters = ["eu-west-1"]
  region      = "eu"
  type        = "service"

  update {
    stagger          = "60s"
    min_healthy_time = "30s"
    healthy_deadline = "2m"
    max_parallel     = 1
    auto_revert      = true
  }

  group "web" {
    count = "{{WEB_TASK_COUNT}}"

    spread {
      attribute = "${node.unique.id}"
      weight    = 100
      # with `target` omitted, Nomad will spread allocations evenly across all values of the attribute.
    }
    spread {
      attribute = "${attr.platform.aws.placement.availability-zone}"
      weight    = 100
      # with `target` omitted, Nomad will spread allocations evenly across all values of the attribute.
    }

    constraint {
      attribute = "${node.class}"
      value     = "web"
    }

    restart {
      attempts = 3
      delay    = "15s"
      interval = "1m"
      mode     = "delay"
    }

    task "dp-frontend-dataset-controller-web" {
      driver = "docker"

      artifact {
        source = "s3::https://s3-eu-west-1.amazonaws.com/{{DEPLOYMENT_BUCKET}}/dp-frontend-dataset-controller/{{REVISION}}.tar.gz"
      }

      config {
        command = "${NOMAD_TASK_DIR}/start-task"

        args = ["./dp-frontend-dataset-controller"]

        image = "{{ECR_URL}}:concourse-{{REVISION}}"

        port_map {
          http = "${NOMAD_PORT_http}"
        }
      }

      service {
        name = "dp-frontend-dataset-controller"
        port = "http"
        tags = ["web"]
        check {
          type     = "http"
          path     = "/health"
          interval = "10s"
          timeout  = "2s"
        }
      }

      resources {
        cpu    = "{{WEB_RESOURCE_CPU}}"
        memory = "{{WEB_RESOURCE_MEM}}"

        network {
          port "http" {}
        }
      }

      template {
        source      = "${NOMAD_TASK_DIR}/vars-template"
        destination = "${NOMAD_TASK_DIR}/vars"
      }

      vault {
        policies = ["dp-frontend-dataset-controller-web"]
      }
    }
  }
  
  group "publishing" {
    count = "{{PUBLISHING_TASK_COUNT}}"

    spread {
      attribute = "${node.unique.id}"
      weight    = 100
      # with `target` omitted, Nomad will spread allocations evenly across all values of the attribute.
    }
    spread {
      attribute = "${attr.platform.aws.placement.availability-zone}"
      weight    = 100
      # with `target` omitted, Nomad will spread allocations evenly across all values of the attribute.
    }
    
    constraint {
      attribute = "${node.class}"
      value     = "publishing"
    }

    restart {
      attempts = 3
      delay    = "15s"
      interval = "1m"
      mode     = "delay"
    }

    task "dp-frontend-dataset-controller-publishing" {
      driver = "docker"

      artifact {
        source = "s3::https://s3-eu-west-1.amazonaws.com/{{DEPLOYMENT_BUCKET}}/dp-frontend-dataset-controller/{{REVISION}}.tar.gz"
      }

      config {
        command = "${NOMAD_TASK_DIR}/start-task"

        args = ["./dp-frontend-dataset-controller"]

        image = "{{ECR_URL}}:concourse-{{REVISION}}"

        port_map {
          http = "${NOMAD_PORT_http}"
        }
      }

      service {
        name = "dp-frontend-dataset-controller"
        port = "http"
        tags = ["publishing"]
        check {
          type     = "http"
          path     = "/health"
          interval = "10s"
          timeout  = "2s"
        }
      }

      resources {
        cpu    = "{{PUBLISHING_RESOURCE_CPU}}"
        memory = "{{PUBLISHING_RESOURCE_MEM}}"

        network {
          port "http" {}
        }
      }

      template {
        source      = "${NOMAD_TASK_DIR}/vars-template"
        destination = "${NOMAD_TASK_DIR}/vars"
      }

      vault {
        policies = ["dp-frontend-dataset-controller-publishing"]
      }
    }
  }
}
