# Service Account for Cloud Scheduler to invoke Cloud Run Jobs
resource "google_service_account" "scheduler_sa" {
  account_id   = "${var.service_name}-scheduler-sa"
  display_name = "Cloud Scheduler Service Account"
}

# IAM binding to allow scheduler to invoke the Cloud Run Jobs
resource "google_project_iam_member" "scheduler_invoke_permission" {
  project = var.project_id
  role    = "roles/run.invoker"
  member  = "serviceAccount:${google_service_account.scheduler_sa.email}"
}

# Morning Push Message Job
resource "google_cloud_run_v2_job" "morning_job" {
  name     = "${var.service_name}-morning-job"
  location = var.region

  template {
    template {
      containers {
        image = "${var.registry_host}/${var.project_id}/${var.service_name}/${var.push_msg_image_name}:latest"

        env {
          name  = "PUSH_MSG_TYPE"
          value = "morning"
        }
        env {
          name  = "PUSH_MSG"
          value = var.morning_msg
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
            cpu    = "1"
            memory = "512Mi"
          }
        }
      }
      vpc_access {
        connector = google_vpc_access_connector.connector.id
        egress    = "ALL_TRAFFIC"
      }
    }
  }

  depends_on = [
    google_sql_user.user,
    google_vpc_access_connector.connector,
    google_sql_database.database
  ]
}

# Evening Push Message Job
resource "google_cloud_run_v2_job" "evening_job" {
  name     = "${var.service_name}-evening-job"
  location = var.region

  template {
    template {
      containers {
        image = "${var.registry_host}/${var.project_id}/${var.service_name}/${var.push_msg_image_name}:latest"

        env {
          name  = "PUSH_MSG_TYPE"
          value = "evening"
        }
        env {
          name  = "PUSH_MSG"
          value = var.evening_msg
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
            cpu    = "1"
            memory = "512Mi"
          }
        }
      }
      vpc_access {
        connector = google_vpc_access_connector.connector.id
        egress    = "ALL_TRAFFIC"
      }
    }
  }

  depends_on = [
    google_sql_user.user,
    google_vpc_access_connector.connector,
    google_sql_database.database
  ]
}

# Cloud Scheduler for Morning Job
resource "google_cloud_scheduler_job" "morning_schedule" {
  name        = "${var.service_name}-morning-schedule"
  description = "Trigger the morning push msg Cloud Run Job"
  schedule    = var.morning_schedule
  time_zone   = "Asia/Taipei"
  region      = var.region

  http_target {
    http_method = "POST"
    uri         = "https://${var.region}-run.googleapis.com/apis/run.googleapis.com/v1/namespaces/${var.project_id}/jobs/${google_cloud_run_v2_job.morning_job.name}:run"
    oauth_token {
      service_account_email = google_service_account.scheduler_sa.email
    }
  }
}

# Cloud Scheduler for Evening Job
resource "google_cloud_scheduler_job" "evening_schedule" {
  name        = "${var.service_name}-evening-schedule"
  description = "Trigger the evening push msg Cloud Run Job"
  schedule    = var.evening_schedule
  time_zone   = "Asia/Taipei"
  region      = var.region

  http_target {
    http_method = "POST"
    uri         = "https://${var.region}-run.googleapis.com/apis/run.googleapis.com/v1/namespaces/${var.project_id}/jobs/${google_cloud_run_v2_job.evening_job.name}:run"
    oauth_token {
      service_account_email = google_service_account.scheduler_sa.email
    }
  }
}
