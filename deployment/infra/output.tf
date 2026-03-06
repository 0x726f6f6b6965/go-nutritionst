output "cloud_run_service_url" {
  description = "The public URL of the deployed Cloud Run service."
  value       = google_cloud_run_v2_service.default.uri
}

output "cloud_sql_instance_connection_name" {
  description = "The connection name for the Cloud SQL instance, which can be used with the Cloud SQL Proxy."
  value       = google_sql_database_instance.postgres_instance.connection_name
}