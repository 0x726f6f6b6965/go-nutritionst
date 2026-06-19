# Build Images

## 1. Prerequisites

- Install [Docker](https://docs.docker.com/desktop/)

## 2. Configure Environment Variables

- Create a .env file in the root directory and add the following variables in the .env file:

```env
DOCKER_HOSTNAME="YOUR_GCP_DOCKER_HOSTNAME"
PROJECT_ID="YOUR_GCP_PROJECT_ID"
IMG_NAME=fastapi-app
```

## 3. Build Images

- Use the following command to build images:

```bash
make build-img
```

or you can build images without CMake:

```bash
docker build -t ${DOCKER_HOSTNAME}/${PROJECT_ID}/go-nutritionst/${IMG_NAME}-line-bot -f ./deployment/line-bot/Dockerfile .
docker build -t ${DOCKER_HOSTNAME}/${PROJECT_ID}/go-nutritionst/${IMG_NAME}-push-msg -f ./deployment/push-msg/Dockerfile .
```