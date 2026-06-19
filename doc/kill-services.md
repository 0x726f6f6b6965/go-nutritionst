# Destroy Services and Infrastructure

This guide explains how to completely destroy and clean up all deployed Google Cloud Platform (GCP) resources using Terraform inside a Docker container.

> [!WARNING]
> Running the destroy commands will permanently delete your infrastructure, including the Cloud Run services, VPC networks, and the Cloud SQL Database (along with all its data). This action cannot be undone.
> If you want to save the data in the database, you can follow the steps on the [backup-db.md](./backup-db.md).

---

## 1. Prerequisites

- Install [Docker Desktop](https://docs.docker.com/desktop/)
- Install [Google Cloud SDK (gcloud CLI)](https://cloud.google.com/sdk/docs/install).

---

## 2. Generate GCP Credentials
Before running the destroy container, ensure you have active credentials on your local machine:

```bash
gcloud auth application-default login
```

---

## 3. Run Destroy

### Option A: Using Makefile (Recommended)
You can build the deployment image and destroy all infrastructure sequentially using the Makefile target:

```bash
make destroy-docker
```

*Note: The script destroys resources in a specific order (Cloud Run -> VPC Connector -> Cloud SQL -> VPC Peering -> VPC Network) to prevent dependency deadlocks and timing issues.*

---

### Option B: Run Docker Directly
If you want to execute the commands directly without using the Makefile:

#### macOS/Linux:
```bash
# Build the deploy image
docker build -t deploy-app -f ./deployment/infra/deploy.Dockerfile .

# Run the sequential destruction
docker run --rm \
  -v ~/.config/gcloud:/root/.config/gcloud \
  -v $(pwd):/workspace \
  deploy-app sh -c " \
    terraform destroy -var-file=terraform.tfvars -var service_name=go-nutritionst -target=google_cloud_run_v2_service.default -auto-approve && \
    terraform destroy -var-file=terraform.tfvars -var service_name=go-nutritionst -target=google_vpc_access_connector.connector -auto-approve && \
    terraform destroy -var-file=terraform.tfvars -var service_name=go-nutritionst -target=google_sql_database.database -auto-approve && \
    terraform destroy -var-file=terraform.tfvars -var service_name=go-nutritionst -target=google_sql_user.user -auto-approve && \
    terraform destroy -var-file=terraform.tfvars -var service_name=go-nutritionst -target=google_sql_database_instance.postgres_instance -auto-approve && \
    terraform destroy -var-file=terraform.tfvars -var service_name=go-nutritionst -target=google_service_networking_connection.private_vpc_connection -auto-approve && \
    terraform destroy -var-file=terraform.tfvars -var service_name=go-nutritionst -target=google_compute_network.vpc_network -auto-approve && \
    terraform destroy -var-file=terraform.tfvars -var service_name=go-nutritionst -auto-approve"
```

#### Windows (PowerShell):
```powershell
# Build the deploy image
docker build -t deploy-app -f ./deployment/infra/deploy.Dockerfile .

# Run the sequential destruction
docker run --rm `
  -v ${HOME}/.config/gcloud:/root/.config/gcloud `
  -v ${PWD}:/workspace `
  deploy-app sh -c " \
    terraform destroy -var-file=terraform.tfvars -var service_name=go-nutritionst -target=google_cloud_run_v2_service.default -auto-approve && \
    terraform destroy -var-file=terraform.tfvars -var service_name=go-nutritionst -target=google_vpc_access_connector.connector -auto-approve && \
    terraform destroy -var-file=terraform.tfvars -var service_name=go-nutritionst -target=google_sql_database.database -auto-approve && \
    terraform destroy -var-file=terraform.tfvars -var service_name=go-nutritionst -target=google_sql_user.user -auto-approve && \
    terraform destroy -var-file=terraform.tfvars -var service_name=go-nutritionst -target=google_sql_database_instance.postgres_instance -auto-approve && \
    terraform destroy -var-file=terraform.tfvars -var service_name=go-nutritionst -target=google_service_networking_connection.private_vpc_connection -auto-approve && \
    terraform destroy -var-file=terraform.tfvars -var service_name=go-nutritionst -target=google_compute_network.vpc_network -auto-approve && \
    terraform destroy -var-file=terraform.tfvars -var service_name=go-nutritionst -auto-approve"
```
