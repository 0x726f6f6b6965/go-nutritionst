# deploy Cloud Run service
resource "google_cloud_run_v2_service" "default" {
  name                = var.service_name
  location            = var.region
  deletion_protection = false
  # Allow unauthenticated public requests
  ingress = "INGRESS_TRAFFIC_ALL"
  template {
    containers {
      image = "${var.registry_host}/${var.project_id}/${var.service_name}/${var.image_name}:latest"
      ports {
        container_port = 8080 # For FastAPI
      }
      env {
        name  = "CHANNEL_ACCESS_TOKEN"
        value = var.channel_access_token
      }

      env {
        name  = "CHANNEL_SECRET"
        value = var.channel_secret
      }

      env {
        name  = "OPENAI_API_KEY"
        value = var.openai_api_key
      }

      env {
        name  = "POSTGRES_PASSWORD"
        value = var.db_pwd
      }

      env {
        name  = "POSTGRES_USER"
        value = google_sql_user.user.name
      }

      env {
        name  = "POSTGRES_DB"
        value = google_sql_database.database.name
      }

      env {
        name  = "POSTGRES_HOST"
        value = google_sql_database_instance.postgres_instance.private_ip_address
      }

      resources {
        limits = {
          cpu    = "2"
          memory = "1024Mi"
        }
      }
    }
    # Connecting Cloud Run to VPC Connector
    vpc_access {
      connector = google_vpc_access_connector.connector.id
      egress    = "ALL_TRAFFIC" # Ensure that all outbound traffic goes through the VPC.
    }
  }

  depends_on = [
    google_sql_user.user,
    google_vpc_access_connector.connector,
    google_sql_database.database
  ]
}


data "google_iam_policy" "noauth" {
  binding {
    role = "roles/run.invoker"
    members = [
      "allUsers",
    ]
  }
}

resource "google_cloud_run_v2_service_iam_policy" "policy" {
  project     = google_cloud_run_v2_service.default.project
  location    = google_cloud_run_v2_service.default.location
  name        = google_cloud_run_v2_service.default.name
  policy_data = data.google_iam_policy.noauth.policy_data
}