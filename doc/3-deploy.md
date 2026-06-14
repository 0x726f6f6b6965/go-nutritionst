# Deploy

## Prepare Terraform Variables

```bash
touch ./deployment/infra/terraform.tfvars
```

## Variables

| variable             | description                                         | default                          |
| -------------------- | --------------------------------------------------- | -------------------------------- |
| region               | The region where resources will be deployed.        | asia-east1                       |
| registry_host        | The hostname of image registry                      | asia-east1-docker.pkg.dev        |
| project_id           | The GCP project ID.                                 |                                  |
| image_name           | The docker image name.                              | fastapi-app-bot                  |
| openai_api_key       | The OpenAI API Key                                  |                                  |
| channel_secret       | The Line OA channel Secret                          |                                  |
| channel_access_token | The Line OA channel access token                    |                                  |
| max_daily_token      | The max daily token                                 | 120000                           |
| db_tier              | Machine type for the Cloud SQL instance.            | db-custom-1-3840                 |
| edition              | The edition of the Cloud SQL instance               | ENTERPRISE                       |
| db_user              | The Cloud SQL database username                     |                                  |
| db_pwd               | The Cloud SQL database password                     |                                  |
| db_name              | Cloud SQL database name                             |                                  |
| push_msg_image_name  | The docker image name for the push message service. | fastapi-app-push-msg             |
| morning_msg          | The afternoon pushing message content               | 早安！記得紀錄今日飲食唷～       |
| evening_msg          | The evening pushing message content                 | 晚安！別忘了確認今日的營養目標！ |
| breakfast_msg        | The breakfast pushing message content               | 早安！記得紀錄今日飲食唷～       |
| lunch_msg            | The lunch pushing message content                   | 午安！記得紀錄今日飲食唷～       |
| dinner_msg           | The dinner pushing message content                  | 晚安！別忘了確認今日的營養目標！ |
| morning_schedule     | Cron schedule for the morning message               | 0 9 * * *                        |
| evening_schedule     | Cron schedule for the evening message               | 0 20 * * *                       |
| breakfast_schedule   | Cron schedule for the breakfast message             | 0 10 * * *                       |
| lunch_schedule       | Cron schedule for the lunch message                 | 0 13 * * *                       |
| dinner_schedule      | Cron schedule for the dinner message                | 0 19 * * *                       |


### Set Variables

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

## Run Deploy

```bash
# 1. Login Google Cloud
gcloud auth login

# 2. Run Deploy
make deploy
```

- After deployment, you will get the public ip from terraform output.
- You need to set the webhook url to Line OA. 
- Example: `https://[IP_ADDRESS]/webhook`

## Run Migration

- Copy `migrations/*.up.sql` and login to Cloud SQL to run migration.

## Line OA Configuration

1. Set Webhook URL

