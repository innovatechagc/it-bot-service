# Makefile para Bot Service

.PHONY: help build run test clean docker-build docker-run docker-test deps lint format swagger deploy-staging deploy-prod sample-data

# Variables
BINARY_NAME=bot-service
DOCKER_IMAGE=bot-service
GO_VERSION=1.21
PROJECT_ID ?= $(shell gcloud config get-value project)
REGION ?= us-central1

help: ## Mostrar ayuda
	@echo "Bot Service - Comandos disponibles:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

deps: ## Instalar dependencias
	go mod download
	go mod tidy

build: ## Compilar la aplicación
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o $(BINARY_NAME) .

run: ## Ejecutar la aplicación
	go run .

test: ## Ejecutar tests
	go test -v ./...

test-coverage: ## Ejecutar tests con cobertura
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

clean: ## Limpiar archivos generados
	go clean
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html

lint: ## Ejecutar linter
	golangci-lint run

format: ## Formatear código
	go fmt ./...
	goimports -w .

swagger: ## Generar documentación Swagger
	swag init

sample-data: ## Crear datos de ejemplo
	go run scripts/sample_data.go

# Docker commands
docker-build: ## Construir imagen Docker
	docker build -t $(DOCKER_IMAGE) .

docker-run: ## Ejecutar contenedor Docker
	docker run -p 8080:8080 --env-file .env.local $(DOCKER_IMAGE)

docker-dev: ## Levantar entorno de desarrollo completo
	docker-compose up -d

docker-down: ## Detener entorno de desarrollo
	docker-compose down

docker-test: ## Ejecutar tests en Docker
	docker-compose -f docker-compose.test.yml up --build --abort-on-container-exit

# GCP Deployment commands
setup-gcp: ## Configurar recursos de GCP
	@echo "Configurando recursos de GCP para bot-service..."
	@./scripts/setup-gcp.sh $(PROJECT_ID) $(REGION)

deploy-staging: ## Deploy a staging usando Cloud Build
	@echo "Desplegando bot-service a staging..."
	@./scripts/deploy.sh staging $(PROJECT_ID) $(REGION)

deploy-prod: ## Deploy a producción usando Cloud Build
	@echo "Desplegando bot-service a producción..."
	@./scripts/deploy.sh production $(PROJECT_ID) $(REGION)

deploy-staging-direct: ## Deploy directo a staging (sin Cloud Build)
	gcloud builds submit --config cloudbuild.yaml \
		--substitutions _ENVIRONMENT=staging,_REGION=$(REGION) \
		--project $(PROJECT_ID)

deploy-prod-direct: ## Deploy directo a producción (sin Cloud Build)
	gcloud builds submit --config cloudbuild.yaml \
		--substitutions _ENVIRONMENT=production,_REGION=$(REGION) \
		--project $(PROJECT_ID)

# Monitoring and maintenance
logs-staging: ## Ver logs de staging
	gcloud run services logs tail bot-service-staging --region=$(REGION) --project=$(PROJECT_ID)

logs-prod: ## Ver logs de producción
	gcloud run services logs tail bot-service-production --region=$(REGION) --project=$(PROJECT_ID)

status-staging: ## Ver estado del servicio en staging
	gcloud run services describe bot-service-staging --region=$(REGION) --project=$(PROJECT_ID)

status-prod: ## Ver estado del servicio en producción
	gcloud run services describe bot-service-production --region=$(REGION) --project=$(PROJECT_ID)

scale-staging: ## Escalar servicio en staging (uso: make scale-staging MIN=1 MAX=5)
	gcloud run services update bot-service-staging \
		--min-instances=$(MIN) --max-instances=$(MAX) \
		--region=$(REGION) --project=$(PROJECT_ID)

scale-prod: ## Escalar servicio en producción (uso: make scale-prod MIN=2 MAX=20)
	gcloud run services update bot-service-production \
		--min-instances=$(MIN) --max-instances=$(MAX) \
		--region=$(REGION) --project=$(PROJECT_ID)

# Testing endpoints
test-staging: ## Probar endpoints en staging
	@echo "Probando bot-service en staging..."
	@STAGING_URL=$$(gcloud run services describe bot-service-staging --region=$(REGION) --project=$(PROJECT_ID) --format='value(status.url)'); \
	curl -f "$$STAGING_URL/api/v1/health" && echo "✓ Health check OK" || echo "✗ Health check failed"; \
	curl -f "$$STAGING_URL/api/v1/ready" && echo "✓ Ready check OK" || echo "✗ Ready check failed"

test-prod: ## Probar endpoints en producción
	@echo "Probando bot-service en producción..."
	@PROD_URL=$$(gcloud run services describe bot-service-production --region=$(REGION) --project=$(PROJECT_ID) --format='value(status.url)'); \
	curl -f "$$PROD_URL/api/v1/health" && echo "✓ Health check OK" || echo "✗ Health check failed"; \
	curl -f "$$PROD_URL/api/v1/ready" && echo "✓ Ready check OK" || echo "✗ Ready check failed"