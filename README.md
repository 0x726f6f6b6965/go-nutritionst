# go-nutritionst

A nutrition assistant bot for Line platform.

## Prerequisites

- Go 1.25.4+
- Python 3.13+
- PostgreSQL
- uv 0.11.7+
- terraform v1.14.8+
- Docker
- gcloud CLI 538.0.0+

### Environment Variables

```bash
touch ./deployment/infra/terraform.tfvars
```

In `terraform.tfvars`, add the following variables:

```hcl
region               = "asia-east1"
registry_host        = "asia-east1-docker.pkg.dev"
project_id           = ""
image_name           = "fastapi-app-bot"
openai_api_key       = ""
channel_access_token = ""
channel_secret       = ""
db_tier              = "db-custom-1-3840"
db_user              = ""
db_pwd               = ""
db_name              = ""
max_daily_token      = "500000"
push_msg_image_name  = "fastapi-app-push-msg"
morning_msg          = "早安！記得紀錄今日睡眠與體重"
evening_msg          = "晚安！別忘了紀錄今日的飲水量與查看每日飲食報告！"
breakfast_msg        = "早安！記得紀錄今日早餐唷～"
lunch_msg            = "午安！記得紀錄今日午餐唷～"
dinner_msg           = "晚安！記得紀錄今日晚餐唷～"
morning_schedule     = "0 9 * * *"
evening_schedule     = "0 20 * * *"
breakfast_schedule   = "0 10 * * *"
lunch_schedule       = "0 13 * * *"
dinner_schedule      = "0 19 * * *"
```
If you want to deploy on local, please follow the below steps:

```bash
touch .env
```

In `.env`, add the following variables:

```text
CHANNEL_ACCESS_TOKEN=""
CHANNEL_SECRET=""
OPENAI_API_KEY=""

DOCKER_HOSTNAME=asia-east1-docker.pkg.dev
PROJECT_ID=""
IMG_NAME=fastapi-app

DATABASE_URL=postgresql://nutritionist:{password}@localhost/postgres

PORT=8080

POSTGRES_DB=nutritionist
POSTGRES_USER=postgres
POSTGRES_PASSWORD=""
POSTGRES_PORT=5432
POSTGRES_HOST=localhost
```

### Build Images and Push to GCR

```bash
make image-push
```

### Upload Manu

Put the image into `./deployment/richmenu/pic/` and name it `example.png`.

```bash
make create-menu
```

### Deployment on GCP

```bash
make deploy
```


