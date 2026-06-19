PROJECTNAME := $(shell basename "$(PWD)")

include .env
export $(shell sed 's/=.*//' .env)

.PHONY: build
build:
	@docker build -t fastapi-app -f ./deployment/line-bot/Dockerfile .

.PHONY: build-broadcast
build-broadcast:
	@docker build -t broadcast-app -f ./deployment/broadcast/Dockerfile .

.PHONY: run
run: build
	@docker run --env-file ./.env -p 8080:80 ${IMG_NAME}

.PHONY: create-menu
create-menu:
	@uv run --env-file .env python ./deployment/richmenu/create_message.py 

.PHONY: build-menu
build-menu:
	@docker build -t richmenu-app -f ./deployment/richmenu/create.Dockerfile .

.PHONY: create-menu-docker
create-menu-docker: build-menu
	@docker run --rm --env-file .env richmenu-app

.PHONY: delete-menu
delete-menu:
	@curl -v -X DELETE \
		https://api.line.me/v2/bot/user/all/richmenu \
		-H "Authorization: Bearer $(CHANNEL_ACCESS_TOKEN)"

.PHONY: build-delete-menu
build-delete-menu:
	@docker build -t delete-menu-app -f ./deployment/richmenu/delete.Dockerfile .

.PHONY: delete-menu-docker
delete-menu-docker: build-delete-menu
	@docker run --rm --env-file .env delete-menu-app

.PHONY: service-up
service-up: build
	@export POSTGRES_DB=nutritionist && \
	export POSTGRES_USER=postgres && \
	export POSTGRES_PASSWORD=docker && \
	export POSTGRES_PORT=5432 && \
	export POSTGRES_HOST=db && \
	docker-compose -f ./deployment/compose.yaml --project-directory . up -d

.PHONY: service-down
service-down:
	@docker-compose -f ./deployment/compose.yaml --project-directory . down 

.PHONY: build-line-bot-img
build-line-bot-img:
	@docker build -t ${DOCKER_HOSTNAME}/${PROJECT_ID}/${PROJECTNAME}/${IMG_NAME}-bot -f ./deployment/line-bot/Dockerfile .

.PHONY: image-push-line-bot
image-push-line-bot:
	@docker push ${DOCKER_HOSTNAME}/${PROJECT_ID}/${PROJECTNAME}/${IMG_NAME}-bot:latest

.PHONY: build-push-msg-img
build-push-msg-img:
	@docker build -t ${DOCKER_HOSTNAME}/${PROJECT_ID}/${PROJECTNAME}/${IMG_NAME}-push-msg -f ./deployment/push-msg/Dockerfile .

.PHONY: image-push-push-msg
image-push-push-msg: 
	@docker push ${DOCKER_HOSTNAME}/${PROJECT_ID}/${PROJECTNAME}/${IMG_NAME}-push-msg:latest


.PHONY: deploy
deploy:
	@terraform -chdir=./deployment/infra apply -var-file="terraform.tfvars" -var service_name=${PROJECTNAME} -auto-approve

.PHONY: build-deploy
build-deploy:
	@docker build -t deploy-app -f ./deployment/infra/deploy.Dockerfile .

.PHONY: deploy-docker
deploy-docker: build-deploy
	@docker run --rm \
		-v ~/.config/gcloud:/root/.config/gcloud \
		-v $(PWD):/workspace \
		deploy-app sh -c "terraform init && terraform apply -var-file=terraform.tfvars -var service_name=${PROJECTNAME} -auto-approve"

.PHONY: destroy-docker
destroy-docker: build-deploy
	@docker run --rm \
		-v ~/.config/gcloud:/root/.config/gcloud \
		-v $(PWD):/workspace \
		deploy-app sh -c "\
			terraform destroy -var-file=terraform.tfvars -var service_name=${PROJECTNAME} -target=google_cloud_run_v2_service.default -auto-approve && \
			terraform destroy -var-file=terraform.tfvars -var service_name=${PROJECTNAME} -target=google_vpc_access_connector.connector -auto-approve && \
			terraform destroy -var-file=terraform.tfvars -var service_name=${PROJECTNAME} -target=google_sql_database.database -auto-approve && \
			terraform destroy -var-file=terraform.tfvars -var service_name=${PROJECTNAME} -target=google_sql_user.user -auto-approve && \
			terraform destroy -var-file=terraform.tfvars -var service_name=${PROJECTNAME} -target=google_sql_database_instance.postgres_instance -auto-approve && \
			terraform destroy -var-file=terraform.tfvars -var service_name=${PROJECTNAME} -target=google_service_networking_connection.private_vpc_connection -auto-approve && \
			terraform destroy -var-file=terraform.tfvars -var service_name=${PROJECTNAME} -target=google_compute_network.vpc_network -auto-approve && \
			terraform destroy -var-file=terraform.tfvars -var service_name=${PROJECTNAME} -auto-approve"

.PHONY: build-migration
build-migration:
	@docker build -t migration-app -f ./deployment/migration/Dockerfile .

.PHONY: migrate-gcp-up
migrate-gcp-up: build-migration
	@docker run --rm \
		-v ~/.config/gcloud:/root/.config/gcloud \
		-e CONNECTION_NAME=$(CONNECTION_NAME) \
		migration-app -path /app/migration -database "postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@127.0.0.1:5432/$(POSTGRES_DB)?sslmode=disable" up

.PHONY: plan
plan:
	@terraform -chdir=./deployment/infra plan -var-file="terraform.tfvars" -var service_name=${PROJECTNAME}

.PHONY: destroy
destroy:
	@terraform -chdir=./deployment/infra destroy -target=google_cloud_run_v2_service.default -auto-approve
	@terraform -chdir=./deployment/infra destroy -target=google_vpc_access_connector.connector -auto-approve
	@terraform -chdir=./deployment/infra destroy -target=google_sql_database.database -auto-approve
	@terraform -chdir=./deployment/infra destroy -target=google_sql_user.user -auto-approve
	@terraform -chdir=./deployment/infra destroy -target=google_sql_database_instance.postgres_instance -auto-approve
	@terraform -chdir=./deployment/infra destroy -target=google_service_networking_connection.private_vpc_connection -auto-approve
	@terraform -chdir=./deployment/infra destroy -target=google_compute_network.vpc_network -auto-approve
	@terraform -chdir=./deployment/infra destroy -auto-approve

.PHONY: tf-fmt
tf-fmt:
	@terraform -chdir=./deployment/infra fmt

