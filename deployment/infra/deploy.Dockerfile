FROM google/cloud-sdk:alpine

# Install Terraform
RUN curl -LO https://releases.hashicorp.com/terraform/1.10.3/terraform_1.10.3_linux_amd64.zip \
    && unzip terraform_1.10.3_linux_amd64.zip -d /usr/bin \
    && rm terraform_1.10.3_linux_amd64.zip

# Set working directory to infra deployment
WORKDIR /workspace/deployment/infra
