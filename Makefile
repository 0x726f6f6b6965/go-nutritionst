PROJECTNAME := $(shell basename "$(PWD)")

include .env
export $(shell sed 's/=.*//' .env)

.PHONY: build
build:
	@docker build -t fastapi-app -f ./deployment/Dockerfile .

.PHONY: run
run: build
	@docker run --env-file ./.env -p 8080:80 ${IMG_NAME}

.PHONY: create-menu
create-menu:
	@uv run --env-file .env python ./deployment/richmenu/create_message.py 

.PHONY: delete-menu
delete-menu:
	@curl -v -X DELETE \
		https://api.line.me/v2/bot/user/all/richmenu \
		-H "Authorization: Bearer $(CHANNEL_ACCESS_TOKEN)"

.PHONY: service-up
service-up: build
	@export POSTGRES_DB=postgres && \
	export POSTGRES_USER=postgres && \
	export POSTGRES_PASSWORD=docker && \
	export POSTGRES_PORT=5432 && \
	export POSTGRES_HOST=db && \
	docker-compose -f ./deployment/compose.yaml --project-directory . up -d

.PHONY: service-down
service-down:
	@docker-compose -f ./deployment/compose.yaml --project-directory . down 

.PHONY: image-push
image-push:build-img
	docker push ${DOCKER_HOSTNAME}/${PROJECT_ID}/${PROJECTNAME}/${IMG_NAME}:latest

.PHONY: build-img
build-img:
	@export TARGETPLATFORM=linux/amd64 && \
	docker build -t ${DOCKER_HOSTNAME}/${PROJECT_ID}/${PROJECTNAME}/${IMG_NAME} -f ./deployment/Dockerfile .


.PHONY: deploy
deploy:
	@terraform -chdir=./deployment/infra apply -var-file="terraform.tfvars" -var service_name=${PROJECTNAME} -auto-approve

.PHONY: plan
plan:
	@terraform -chdir=./deployment/infra plan -var-file="terraform.tfvars" -var service_name=${PROJECTNAME}

.PHONY: destroy
destroy:
	@terraform -chdir=./deployment/infra destroy -target=google_cloud_run_v2_service.default -auto-approve
	@terraform -chdir=./deployment/infra destroy -target=google_vpc_access_connector.connector -auto-approve
	@terraform -chdir=./deployment/infra destroy -target=google_sql_database_instance.postgres_instance -auto-approve
	@terraform -chdir=./deployment/infra destroy -target=google_service_networking_connection.private_vpc_connection -auto-approve
	@terraform -chdir=./deployment/infra destroy -target=google_compute_network.vpc_network -auto-approve
	@terraform -chdir=./deployment/infra destroy -auto-approve

.PHONY: tf-fmt
tf-fmt:
	@terraform -chdir=./deployment/infra fmt

