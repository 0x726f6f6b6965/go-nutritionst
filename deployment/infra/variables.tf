variable "region" {
  description = "The region where resources will be deployed."
  type        = string
  default     = "asia-east2"
}

variable "registry_host" {
  description = "The hostname of image registry"
  type        = string
  default     = "registry.local"
}

variable "project_id" {
  description = "The GCP project ID."
  type        = string
  default     = "project-id"
}

variable "service_name" {
  description = "The name of the repository. It will be the base name for Cloud Run and other resources."
  type        = string
  default     = "repo"
}

variable "image_name" {
  description = "The docker image name."
  type        = string
  default     = "image-name"
}

variable "openai_api_key" {
  description = "The OpenAI API Key"
  type        = string
  default     = "OPENAI_API_KEY"
  sensitive   = true
}

variable "channel_secret" {
  description = "The Line OA channel Secret"
  type        = string
  default     = "CHANNEL_SECRET"
  sensitive   = true
}

variable "channel_access_token" {
  description = "The Line OA channel access token"
  type        = string
  default     = "CHANNEL_ACCESS_TOKEN"
  sensitive   = true
}

variable "db_tier" {
  description = "Machine type for the Cloud SQL instance."
  type        = string
  default     = "db-custom-1-3840"
}

variable "edition" {
  description = "The edition of the Cloud SQL instance"
  type        = string
  default     = "ENTERPRISE"
}

variable "db_user" {
  description = "The Cloud SQL database username"
  type        = string
  default     = "db-user"
}

variable "db_pwd" {
  description = "The Cloud SQL database password"
  type        = string
  default     = "db-pwd"
  sensitive   = true
}

variable "db_name" {
  description = "Cloud SQL database name"
  type        = string
  default     = "fastapi_db"
}