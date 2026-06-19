# Upload Images to GCP Artifact Registry

This guide explains how to authenticate with Google Cloud Platform (GCP), configure your environment, and build/push Docker images to GCP Artifact Registry.

---

## 1. Prerequisites

- Install [Docker Desktop](https://docs.docker.com/desktop/) (which includes Docker Buildx).
- Install [Google Cloud SDK (gcloud CLI)](https://cloud.google.com/sdk/docs/install).
- Run the [build-images.md](./build-images.md) to build images first.

---

## 2. GCP Authentication & Repository Setup

### A. Authenticate with GCP
Log in to your Google Cloud account via the CLI:
```bash
gcloud auth login
```

Set your active project:
```bash
gcloud config set project YOUR_GCP_PROJECT_ID
```

### B. Create an Artifact Registry Repository (If not already created)
Run the following command to create a Docker repository. Replace `<region>` with your preferred GCP region (e.g., `asia-east1`) and `<repository-name>` with your repository name (e.g., `go-nutritionist`):
```bash
gcloud artifacts repositories create <repository-name> \
    --repository-format=docker \
    --location=<region> \
    --description="Docker repository for Nutritionist App"
```

### C. Configure Docker Authentication
Configure Docker to authenticate requests to the Artifact Registry for your region:
```bash
gcloud auth configure-docker <region>-docker.pkg.dev
```

---

## 3. Configure Environment Variables

Create (or update) the `.env` file in the root directory and ensure these variables are set correctly:

```env
DOCKER_HOSTNAME="<region>-docker.pkg.dev"
PROJECT_ID="YOUR_GCP_PROJECT_ID"
IMG_NAME="fastapi-app"
```

---

## 4. Push Docker Images

- Use the following command to upload the images to GCP Artifact Registry:

```bash
make image-push-line-bot
make image-push-push-msg
```

or you can use the following command to upload the images to GCP Artifact Registry without CMake:

```bash
docker push ${DOCKER_HOSTNAME}/${PROJECT_ID}/go-nutritionst/${IMG_NAME}-bot:latest
docker push ${DOCKER_HOSTNAME}/${PROJECT_ID}/go-nutritionst/${IMG_NAME}-push-msg:latest
```
