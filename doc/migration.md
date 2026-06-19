# Database Migration on GCP

This guide explains how to build a Docker container that bundles `golang-migrate` and `cloud-sql-proxy` to securely execute database migrations against your GCP Cloud SQL (PostgreSQL) instance.

---

## 1. Prerequisites

- Install [Docker Desktop](https://docs.docker.com/desktop/)
- Install [Google Cloud SDK (gcloud CLI)](https://cloud.google.com/sdk/docs/install).
- Run the [deploy-services.md](./deploy-services.md) to deploy the infrastructure first.

---

## 2. Generate GCP Credentials
Before running the migration container, ensure you have generated **Application Default Credentials (ADC)** on your local host machine. The container will mount these credentials to authenticate the Cloud SQL Auth Proxy.

```bash
gcloud auth application-default login
```

---

## 3. Configure Environment Variables
Ensure your `.env` file in the root directory contains the following database variables and the GCP Cloud SQL Connection Name:

```env
POSTGRES_USER="postgres"
POSTGRES_PASSWORD="YOUR_DATABASE_PASSWORD"
POSTGRES_DB="nutritionist"

# GCP Cloud SQL Connection Name (format: PROJECT_ID:REGION:INSTANCE_NAME)
CONNECTION_NAME="YOUR_GCP_PROJECT_ID:asia-east1:go-nutritionst-db-instance"
```

---

## 4. Run Migrations

### Option A: Using Makefile (Recommended)
You can build the migration image and run all migrations up to the latest version with one command:

```bash
make migrate-gcp-up
```

### Option B: Run Docker Directly
If you want to run `golang-migrate` commands manually (e.g. `up`, `down`, `force`, etc.), you can run the Docker command directly:

#### Build the image:
```bash
docker build -t migration-app -f ./deployment/migration/Dockerfile .
```

#### Run migrate up:
```bash
docker run --rm \
  -v ~/.config/gcloud:/root/.config/gcloud \
  -e CONNECTION_NAME="YOUR_GCP_PROJECT_ID:asia-east1:go-nutritionst-db-instance" \
  migration-app -path /app/migration -database "postgres://postgres:YOUR_DATABASE_PASSWORD@127.0.0.1:5432/nutritionist?sslmode=disable" up
```

#### Run migrate down (Rollback):
```bash
docker run --rm \
  -v ~/.config/gcloud:/root/.config/gcloud \
  -e CONNECTION_NAME="YOUR_GCP_PROJECT_ID:asia-east1:go-nutritionst-db-instance" \
  migration-app -path /app/migration -database "postgres://postgres:YOUR_DATABASE_PASSWORD@127.0.0.1:5432/nutritionist?sslmode=disable" down
```

#### Run migrate force (Fix dirty migration state):
```bash
docker run --rm \
  -v ~/.config/gcloud:/root/.config/gcloud \
  -e CONNECTION_NAME="YOUR_GCP_PROJECT_ID:asia-east1:go-nutritionst-db-instance" \
  migration-app -path /app/migration -database "postgres://postgres:YOUR_DATABASE_PASSWORD@127.0.0.1:5432/nutritionist?sslmode=disable" force 1
```

---

## How It Works Under the Hood
1. The **Cloud SQL Auth Proxy** starts inside the container and opens a local port (`127.0.0.1:5432`).
2. It secures the connection using the mounted Google credentials from your host (`~/.config/gcloud`).
3. The **golang-migrate CLI** runs and connects to `127.0.0.1:5432`, forwarding the traffic securely to Cloud SQL on GCP.
