#!/bin/bash

# GCP Setup Script for Bot Service
# Usage: ./scripts/setup-gcp.sh [PROJECT_ID] [REGION]

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default values
PROJECT_ID=${1:-""}
REGION=${2:-us-central1}
REPOSITORY="bot-service-repo"
SERVICE_ACCOUNT="bot-service"

# Get project ID if not provided
if [[ -z "$PROJECT_ID" ]]; then
    PROJECT_ID=$(gcloud config get-value project 2>/dev/null)
    if [[ -z "$PROJECT_ID" ]]; then
        echo -e "${RED}Error: PROJECT_ID not found. Please provide it as first argument or set gcloud config${NC}"
        exit 1
    fi
fi

echo -e "${BLUE}🚀 Setting up GCP resources for bot-service${NC}"
echo -e "${BLUE}Project: ${PROJECT_ID}${NC}"
echo -e "${BLUE}Region: ${REGION}${NC}"

# Enable required APIs
echo -e "${YELLOW}📋 Enabling required APIs...${NC}"
REQUIRED_APIS=(
    "cloudbuild.googleapis.com"
    "run.googleapis.com"
    "artifactregistry.googleapis.com"
    "secretmanager.googleapis.com"
    "compute.googleapis.com"
    "vpcaccess.googleapis.com"
    "redis.googleapis.com"
    "sqladmin.googleapis.com"
)

for api in "${REQUIRED_APIS[@]}"; do
    echo -e "${YELLOW}Enabling $api...${NC}"
    gcloud services enable "$api" --project="$PROJECT_ID"
done

# Create Artifact Registry repository
echo -e "${YELLOW}📦 Creating Artifact Registry repository...${NC}"
if ! gcloud artifacts repositories describe "$REPOSITORY" --location="$REGION" --project="$PROJECT_ID" >/dev/null 2>&1; then
    gcloud artifacts repositories create "$REPOSITORY" \
        --repository-format=docker \
        --location="$REGION" \
        --description="Bot Service Docker repository" \
        --project="$PROJECT_ID"
    echo -e "${GREEN}✓ Artifact Registry repository created${NC}"
else
    echo -e "${GREEN}✓ Artifact Registry repository already exists${NC}"
fi

# Create service account
echo -e "${YELLOW}🔐 Creating service account...${NC}"
if ! gcloud iam service-accounts describe "${SERVICE_ACCOUNT}@${PROJECT_ID}.iam.gserviceaccount.com" --project="$PROJECT_ID" >/dev/null 2>&1; then
    gcloud iam service-accounts create "$SERVICE_ACCOUNT" \
        --display-name="Bot Service Account" \
        --description="Service account for bot-service Cloud Run deployment" \
        --project="$PROJECT_ID"
    echo -e "${GREEN}✓ Service account created${NC}"
else
    echo -e "${GREEN}✓ Service account already exists${NC}"
fi

# Grant necessary IAM roles to service account
echo -e "${YELLOW}🔑 Granting IAM roles to service account...${NC}"
ROLES=(
    "roles/cloudsql.client"
    "roles/secretmanager.secretAccessor"
    "roles/redis.editor"
    "roles/logging.logWriter"
    "roles/monitoring.metricWriter"
    "roles/cloudtrace.agent"
)

for role in "${ROLES[@]}"; do
    gcloud projects add-iam-policy-binding "$PROJECT_ID" \
        --member="serviceAccount:${SERVICE_ACCOUNT}@${PROJECT_ID}.iam.gserviceaccount.com" \
        --role="$role" \
        --quiet
done

echo -e "${GREEN}✓ IAM roles granted${NC}"

# Create VPC connector (if it doesn't exist)
echo -e "${YELLOW}🌐 Creating VPC connector...${NC}"
CONNECTOR_NAME="bot-service-connector"
if ! gcloud compute networks vpc-access connectors describe "$CONNECTOR_NAME" --region="$REGION" --project="$PROJECT_ID" >/dev/null 2>&1; then
    gcloud compute networks vpc-access connectors create "$CONNECTOR_NAME" \
        --region="$REGION" \
        --subnet-project="$PROJECT_ID" \
        --subnet="default" \
        --range="10.8.0.0/28" \
        --min-instances=2 \
        --max-instances=10 \
        --project="$PROJECT_ID"
    echo -e "${GREEN}✓ VPC connector created${NC}"
else
    echo -e "${GREEN}✓ VPC connector already exists${NC}"
fi

# Create secrets (with placeholder values)
echo -e "${YELLOW}🔑 Creating secrets...${NC}"
SECRETS=(
    "openai-api-key:sk-placeholder-openai-key"
    "database-password-staging:staging-db-password"
    "database-password-production:production-db-password"
    "redis-password-staging:staging-redis-password"
    "redis-password-production:production-redis-password"
    "jwt-secret-production:jwt-secret-placeholder"
)

for secret_info in "${SECRETS[@]}"; do
    secret_name=$(echo "$secret_info" | cut -d':' -f1)
    secret_value=$(echo "$secret_info" | cut -d':' -f2)
    
    if ! gcloud secrets describe "$secret_name" --project="$PROJECT_ID" >/dev/null 2>&1; then
        echo "$secret_value" | gcloud secrets create "$secret_name" \
            --data-file=- \
            --project="$PROJECT_ID"
        echo -e "${GREEN}✓ Secret '$secret_name' created with placeholder value${NC}"
        echo -e "${YELLOW}⚠️  Remember to update the secret value: gcloud secrets versions add $secret_name --data-file=-${NC}"
    else
        echo -e "${GREEN}✓ Secret '$secret_name' already exists${NC}"
    fi
done

# Create Cloud SQL instance (optional - commented out by default)
echo -e "${YELLOW}💾 Cloud SQL setup (optional)...${NC}"
echo -e "${BLUE}To create a Cloud SQL instance, run:${NC}"
echo "gcloud sql instances create bot-service-db \\"
echo "  --database-version=POSTGRES_14 \\"
echo "  --tier=db-f1-micro \\"
echo "  --region=$REGION \\"
echo "  --storage-type=SSD \\"
echo "  --storage-size=10GB \\"
echo "  --project=$PROJECT_ID"

# Create Redis instance (optional - commented out by default)
echo -e "${YELLOW}🔴 Redis setup (optional)...${NC}"
echo -e "${BLUE}To create a Redis instance, run:${NC}"
echo "gcloud redis instances create bot-service-redis \\"
echo "  --size=1 \\"
echo "  --region=$REGION \\"
echo "  --redis-version=redis_6_x \\"
echo "  --project=$PROJECT_ID"

# Configure Docker authentication
echo -e "${YELLOW}🐳 Configuring Docker authentication...${NC}"
gcloud auth configure-docker "${REGION}-docker.pkg.dev" --quiet
echo -e "${GREEN}✓ Docker authentication configured${NC}"

echo -e "${GREEN}✅ GCP setup completed successfully!${NC}"
echo -e "${GREEN}Resources created:${NC}"
echo -e "${GREEN}  - Artifact Registry repository: ${REPOSITORY}${NC}"
echo -e "${GREEN}  - Service account: ${SERVICE_ACCOUNT}@${PROJECT_ID}.iam.gserviceaccount.com${NC}"
echo -e "${GREEN}  - VPC connector: ${CONNECTOR_NAME}${NC}"
echo -e "${GREEN}  - Secrets: openai-api-key, database-password-*, redis-password-*, jwt-secret-production${NC}"

echo -e "${BLUE}📋 Next steps:${NC}"
echo -e "${BLUE}1. Update secret values with real credentials${NC}"
echo -e "${BLUE}2. Create Cloud SQL and Redis instances if needed${NC}"
echo -e "${BLUE}3. Run deployment: make deploy-staging${NC}"

echo -e "${YELLOW}⚠️  Important: Update the placeholder secret values before deploying!${NC}"