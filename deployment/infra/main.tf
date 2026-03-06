# Configuration
terraform {
  required_version = ">=1.0"
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = ">= 4.40.0"
    }
    google-beta = {
      source  = "hashicorp/google-beta"
      version = ">= 4.50.0"
    }
  }
}

# Provider
provider "google" {
  project = var.project_id
  region  = var.region
}

provider "google-beta" {
  project = var.project_id
  region  = var.region
}

resource "google_project_service" "apis" {
  for_each = toset([
    "run.googleapis.com",             # Cloud Run API
    "sqladmin.googleapis.com",        # Cloud SQL Admin API
    "compute.googleapis.com",         # Compute Engine API (for VPC)
    "vpcaccess.googleapis.com",       # Serverless VPC Access API
    "artifactregistry.googleapis.com" # Artifact Registry API
  ])
  service            = each.key
  disable_on_destroy = false
}