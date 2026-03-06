# Create a VPC network for Cloud SQL
resource "google_compute_network" "vpc_network" {
  name                    = "${var.service_name}-vpc"
  auto_create_subnetworks = false # Manually manage subnet routes
  depends_on = [
    google_project_service.apis
  ]
}

# Create a Cloud Router
# Cloud NAT requires a Cloud Router to function.
resource "google_compute_router" "router" {
  name    = "${var.service_name}-router"
  region  = var.region
  network = google_compute_network.vpc_network.id
}

# Create a Cloud NAT
# This will allow resources within the VPC to access the public internet.
resource "google_compute_router_nat" "nat" {
  name                               = "${var.service_name}-nat"
  router                             = google_compute_router.router.name
  region                             = google_compute_router.router.region
  source_subnetwork_ip_ranges_to_nat = "ALL_SUBNETWORKS_ALL_IP_RANGES"
  nat_ip_allocate_option             = "AUTO_ONLY"

  log_config {
    enable = true
    filter = "ERRORS_ONLY"
  }
}

# Add a subnetwork to the VPC
resource "google_compute_subnetwork" "subnet" {
  name          = "${var.service_name}-subnet"
  ip_cidr_range = "10.0.1.0/24"
  region        = var.region # Change this to your desired region
  network       = google_compute_network.vpc_network.self_link
  depends_on = [
    google_project_service.apis,
    google_compute_network.vpc_network
  ]
}

# Create a global  address for private service connection
resource "google_compute_global_address" "private_ip_address" {
  provider      = google-beta
  name          = "${var.service_name}-private-ip"
  purpose       = "VPC_PEERING"
  address_type  = "INTERNAL"
  prefix_length = 16
  network       = google_compute_network.vpc_network.id
}

# Create Private Service Connection
resource "google_service_networking_connection" "private_vpc_connection" {
  provider                = google-beta
  network                 = google_compute_network.vpc_network.id
  service                 = "servicenetworking.googleapis.com"
  reserved_peering_ranges = [google_compute_global_address.private_ip_address.name]
  deletion_policy         = "ABANDON"
  depends_on = [
    google_project_service.apis,
    google_compute_network.vpc_network
  ]
}

# Create a Serverless VPC Access Connector
resource "google_vpc_access_connector" "connector" {
  name           = "${var.service_name}-connector"
  region         = var.region
  ip_cidr_range  = "10.8.0.0/28" # This IP address range cannot overlap with your VPC subnet range.
  network        = google_compute_network.vpc_network.name
  machine_type   = "f1-micro"
  min_throughput = 200 # Valid value
  max_throughput = 1000

  depends_on = [
    google_project_service.apis,
    google_compute_network.vpc_network
  ]
}