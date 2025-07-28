#!/bin/bash

# Bot Service Deployment Script for GCP Cloud Run
# Usage: ./scripts/deploy.sh [staging|production] [PROJECT_ID] [REGION]

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default values
ENVIRONMENT=${1:-staging}
PROJECT_ID=${2:-""}
REGION=${3:-us-central1}
REPOSITORY="bot-service-repo"
SERVICE_NAME="bot-service-${ENVIRONMENT}"

# Validate environment
if [[ "$ENVIRONMENT" != "staging" && "$ENVIRONMENT" != "production" ]]; then
    echo -e "${RED}Error: Environment must be 'staging' or 'production'${NC}"
    echo "Usage: $0 [staging|production] [PROJECT_ID] [REGION]"
    exit 1
fi

# Get project ID if not provided
if [[ -z "$PROJECT_ID" ]]; then
    PROJECT_ID=$(gcloud config get-value project 2>/dev/null)
    if [[ -z "$PROJECT_ID" ]]; then
        echo -e "${RED}Error: PROJECT_ID not found. Please provide it as second argument or set gcloud config${NC}"
        exit 1
    fi
fi

echo -e "${BLUE}🚀 Starting deployment of bot-service to ${ENVIRONMENT}${NC}"
echo -e "${BLUE}Project: ${PROJECT_ID}${NC}"
echo -e "${BLUE}Region: ${REGION}${NC}"
echo -e "${BLUE}Repository: ${REPOSITORY}${NC}"

# Set environment-specific variables
if [[ "$ENVIRONMENT" == "production" ]]; then
    MEMORY="2Gi"
    CPU="2"
    MIN_INSTANCES="2"
    MAX_INSTANCES="20"
    CONCURRENCY="100"
    LOG_LEVEL="warn"
else
    MEMORY="1Gi"
    CPU="1"
    MIN_INSTANCES="0"
    MAX_INSTANCES="5"
    CONCURRENCY="80"
    LOG_LEVEL="debug"
fi

# Check if required APIs are enabled
echo -e "${YELLOW}📋 Checking required APIs...${NC}"
REQUIRED_APIS=(
    "cloudbuild.googleapis.com"
    "run.googleapis.com"
    "artifactregistry.googleapis.com"
    "secretmanager.googleapis.com"
)

for api in "${REQUIRED_APIS[@]}"; do
    if ! gcloud services list --enabled --filter="name:$api" --format="value(name)" | grep -q "$api"; then
        echo -e "${YELLOW}Enabling $api...${NC}"
        gcloud services enable "$api" --project="$PROJECT_ID"
    else
        echo -e "${GREEN}✓ $api is enabled${NC}"
    fi
done

# Create Artifact Registry repository if it doesn't exist
echo -e "${YELLOW}📦 Checking Artifact Registry repository...${NC}"
if ! gcloud artifacts repositories describe "$REPOSITORY" --location="$REGION" --project="$PROJECT_ID" >/dev/null 2>&1; then
    echo -e "${YELLOW}Creating Artifact Registry repository...${NC}"
    gcloud artifacts repositories create "$REPOSITORY" \
        --repository-format=docker \
        --location="$REGION" \
        --description="Bot Service Docker repository" \
        --project="$PROJECT_ID"
else
    echo -e "${GREEN}✓ Artifact Registry repository exists${NC}"
fi

# Configure Docker authentication
echo -e "${YELLOW}🔐 Configuring Docker authentication...${NC}"
gcloud auth configure-docker "${REGION}-docker.pkg.dev" --quiet

# Check if required secrets exist
echo -e "${YELLOW}🔑 Checking required secrets...${NC}"
REQUIRED_SECRETS=(
    "openai-api-key"
    "database-password-${ENVIRONMENT}"
    "redis-password-${ENVIRONMENT}"
)

if [[ "$ENVIRONMENT" == "production" ]]; then
    REQUIRED_SECRETS+=("jwt-secret-production")
fi

for secret in "${REQUIRED_SECRETS[@]}"; do
    if ! gcloud secrets describe "$secret" --project="$PROJECT_ID" >/dev/null 2>&1; then
        echo -e "${RED}⚠️  Secret '$secret' not found. Please create it first:${NC}"
        echo "gcloud secrets create $secret --data-file=- --project=$PROJECT_ID"
        echo "Then add the secret value when prompted."
    else
        echo -e "${GREEN}✓ Secret '$secret' exists${NC}"
    fi
done

# Build and deploy using Cloud Build
echo -e "${YELLOW}🏗️  Starting Cloud Build...${NC}"
gcloud builds submit \
    --config=cloudbuild.yaml \
    --substitutions="\
_REGION=${REGION},\
_REPOSITORY=${REPOSITORY},\
_ENVIRONMENT=${ENVIRONMENT},\
_MEMORY=${MEMORY},\
_CPU=${CPU},\
_MIN_INSTANCES=${MIN_INSTANCES},\
_MAX_INSTANCES=${MAX_INSTANCES},\
_CONCURRENCY=${CONCURRENCY},\
_LOG_LEVEL=${LOG_LEVEL},\
_OPENAI_SECRET=openai-api-key,\
_DB_SECRET=database-password-${ENVIRONMENT},\
_VPC_CONNECTOR=projects/${PROJECT_ID}/locations/${REGION}/connectors/bot-service-connector,\
_SERVICE_ACCOUNT=bot-service@${PROJECT_ID}.iam.gserviceaccount.com" \
    --project="$PROJECT_ID"

# Get service URL
SERVICE_URL=$(gcloud run services describe "$SERVICE_NAME" \
    --region="$REGION" \
    --project="$PROJECT_ID" \
    --format='value(status.url)')

echo -e "${GREEN}✅ Deployment completed successfully!${NC}"
echo -e "${GREEN}Service URL: ${SERVICE_URL}${NC}"
echo -e "${GREEN}Health Check: ${SERVICE_URL}/api/v1/health${NC}"
echo -e "${GREEN}API Documentation: ${SERVICE_URL}/swagger/index.html${NC}"

# Run post-deployment tests
echo -e "${YELLOW}🧪 Running post-deployment tests...${NC}"
sleep 10

# Health check
if curl -f -s "${SERVICE_URL}/api/v1/health" >/dev/null; then
    echo -e "${GREEN}✓ Health check passed${NC}"
else
    echo -e "${RED}✗ Health check failed${NC}"
    exit 1
fi

# Readiness check
if curl -f -s "${SERVICE_URL}/api/v1/ready" >/dev/null; then
    echo -e "${GREEN}✓ Readiness check passed${NC}"
else
    echo -e "${RED}✗ Readiness check failed${NC}"
    exit 1
fi

# Test bot endpoint
if curl -f -s "${SERVICE_URL}/api/v1/bots" >/dev/null; then
    echo -e "${GREEN}✓ Bot API endpoint accessible${NC}"
else
    echo -e "${YELLOW}⚠️  Bot API endpoint test failed (may require authentication)${NC}"
fi

echo -e "${GREEN}🎉 Deployment and tests completed successfully!${NC}"

# Show useful commands
echo -e "${BLUE}📋 Useful commands:${NC}"
echo "View logs: gcloud run services logs tail $SERVICE_NAME --region=$REGION --project=$PROJECT_ID"
echo "Update traffic: gcloud run services update-traffic $SERVICE_NAME --to-latest --region=$REGION --project=$PROJECT_ID"
echo "Scale service: gcloud run services update $SERVICE_NAME --min-instances=N --max-instances=N --region=$REGION --project=$PROJECT_ID"