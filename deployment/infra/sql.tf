# Create a Cloud SQL (PostgreSQL) instance
resource "google_sql_database_instance" "postgres_instance" {
  database_version    = "POSTGRES_15"
  region              = var.region
  name                = "${var.service_name}-db-instance"
  deletion_protection = false
  settings {
    tier    = var.db_tier
    edition = var.edition
    backup_configuration {
      enabled = true
    }
    ip_configuration {
      ipv4_enabled    = false
      private_network = google_compute_network.vpc_network.id
    }
  }

  depends_on = [
    google_service_networking_connection.private_vpc_connection
  ]
}


# Create a database on Cloud SQL
resource "google_sql_database" "database" {
  name     = var.db_name
  instance = google_sql_database_instance.postgres_instance.name
  charset  = "UTF8"
}

# Create a database user
resource "google_sql_user" "user" {
  name     = var.db_user
  instance = google_sql_database_instance.postgres_instance.name
  password = var.db_pwd
}