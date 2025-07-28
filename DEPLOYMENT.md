# Bot Service - Guía de Deployment en GCP

Esta guía te ayudará a desplegar el bot-service en Google Cloud Platform usando Cloud Run.

## 📋 Prerrequisitos

1. **Google Cloud SDK** instalado y configurado
2. **Docker** instalado
3. **Proyecto de GCP** creado y configurado
4. **Permisos necesarios** en el proyecto GCP:
   - Cloud Build Editor
   - Cloud Run Admin
   - Artifact Registry Admin
   - Secret Manager Admin
   - Service Account Admin

## 🚀 Deployment Rápido

### 1. Configuración Inicial

```bash
# Configurar proyecto GCP
gcloud config set project YOUR_PROJECT_ID

# Ejecutar setup automático de recursos GCP
make setup-gcp

# O manualmente:
./scripts/setup-gcp.sh YOUR_PROJECT_ID us-central1
```

### 2. Actualizar Secretos

Después del setup inicial, actualiza los secretos con valores reales:

```bash
# OpenAI API Key
echo "sk-your-real-openai-key" | gcloud secrets versions add openai-api-key --data-file=-

# Database passwords
echo "your-staging-db-password" | gcloud secrets versions add database-password-staging --data-file=-
echo "your-production-db-password" | gcloud secrets versions add database-password-production --data-file=-

# Redis passwords
echo "your-staging-redis-password" | gcloud secrets versions add redis-password-staging --data-file=-
echo "your-production-redis-password" | gcloud secrets versions add redis-password-production --data-file=-

# JWT secret (solo para producción)
echo "your-jwt-secret" | gcloud secrets versions add jwt-secret-production --data-file=-
```

### 3. Deploy a Staging

```bash
# Usando el script automatizado
make deploy-staging

# O directamente con Cloud Build
make deploy-staging-direct
```

### 4. Deploy a Producción

```bash
# Usando el script automatizado
make deploy-prod

# O directamente con Cloud Build
make deploy-prod-direct
```

## 🔧 Configuración Detallada

### Variables de Entorno por Ambiente

#### Staging
- `ENVIRONMENT=staging`
- `LOG_LEVEL=debug`
- `MIN_INSTANCES=0`
- `MAX_INSTANCES=5`
- `MEMORY=1Gi`
- `CPU=1`

#### Production
- `ENVIRONMENT=production`
- `LOG_LEVEL=warn`
- `MIN_INSTANCES=2`
- `MAX_INSTANCES=20`
- `MEMORY=2Gi`
- `CPU=2`

### Personalizar Deployment

Puedes personalizar el deployment modificando las variables de sustitución en `cloudbuild.yaml`:

```bash
gcloud builds submit --config cloudbuild.yaml \
  --substitutions _ENVIRONMENT=staging,_REGION=us-central1,_MEMORY=2Gi,_CPU=2
```

## 📊 Monitoreo y Mantenimiento

### Ver Logs

```bash
# Staging
make logs-staging

# Production
make logs-prod
```

### Verificar Estado

```bash
# Staging
make status-staging

# Production
make status-prod
```

### Escalar Servicios

```bash
# Staging
make scale-staging MIN=1 MAX=10

# Production
make scale-prod MIN=5 MAX=50
```

### Probar Endpoints

```bash
# Staging
make test-staging

# Production
make test-prod
```

## 🏗️ Arquitectura del Deployment

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   Cloud Build   │───▶│ Artifact Registry│───▶│   Cloud Run     │
│                 │    │                  │    │                 │
│ - Run Tests     │    │ - Docker Images  │    │ - bot-service   │
│ - Build Image   │    │ - Version Tags   │    │ - Auto-scaling  │
│ - Deploy        │    │                  │    │ - Health Checks │
└─────────────────┘    └──────────────────┘    └─────────────────┘
         │                                               │
         │              ┌──────────────────┐            │
         └─────────────▶│ Secret Manager   │◀───────────┘
                        │                  │
                        │ - API Keys       │
                        │ - DB Passwords   │
                        │ - JWT Secrets    │
                        └──────────────────┘
```

## 🔐 Seguridad

### Service Account
El servicio usa una service account dedicada con permisos mínimos:
- `roles/cloudsql.client`
- `roles/secretmanager.secretAccessor`
- `roles/redis.editor`
- `roles/logging.logWriter`
- `roles/monitoring.metricWriter`

### Secretos
Todos los secretos sensibles se almacenan en Secret Manager:
- API keys de servicios externos
- Passwords de bases de datos
- JWT secrets
- Credenciales de Redis

### Red
- VPC Connector para acceso privado a recursos internos
- Egress solo a rangos privados
- HTTPS obligatorio

## 🧪 Testing

### Tests Locales

```bash
# Ejecutar tests
make test

# Tests con cobertura
make test-coverage

# Crear datos de ejemplo
make sample-data
```

### Tests Post-Deployment

Los tests automáticos incluyen:
- Health check (`/api/v1/health`)
- Readiness check (`/api/v1/ready`)
- Conectividad básica de APIs

## 📈 Endpoints Disponibles

Una vez desplegado, el servicio expone:

### Health & Monitoring
- `GET /api/v1/health` - Estado del servicio
- `GET /api/v1/ready` - Readiness check
- `GET /metrics` - Métricas de Prometheus

### Bot Management
- `GET /api/v1/bots` - Lista de bots
- `POST /api/v1/bots` - Crear bot
- `GET /api/v1/bots/:id` - Detalle de bot
- `PATCH /api/v1/bots/:id` - Actualizar bot
- `DELETE /api/v1/bots/:id` - Eliminar bot

### Flow Management
- `GET /api/v1/bots/:id/flows` - Flujos del bot
- `POST /api/v1/bots/:id/flows` - Crear flujo
- `GET /api/v1/flows/:id` - Detalle de flujo
- `PATCH /api/v1/flows/:id` - Actualizar flujo
- `DELETE /api/v1/flows/:id` - Eliminar flujo

### Step Management
- `POST /api/v1/flows/:id/steps` - Crear paso
- `PATCH /api/v1/steps/:id` - Actualizar paso
- `DELETE /api/v1/steps/:id` - Eliminar paso

### AI & Smart Replies
- `POST /api/v1/bots/:id/smart-reply` - Consulta a IA
- `POST /api/v1/bots/:id/intents/train` - Entrenar intents
- `GET /api/v1/bots/:id/intents` - Lista de intents

### Message Processing
- `POST /api/v1/incoming` - Procesar mensaje entrante

### Documentación
- `GET /swagger/index.html` - Documentación Swagger completa

## 🚨 Troubleshooting

### Errores Comunes

1. **Build fails**: Verificar que todas las APIs estén habilitadas
2. **Secrets not found**: Asegurar que todos los secretos existan
3. **VPC connector issues**: Verificar configuración de red
4. **Service account permissions**: Revisar roles IAM

### Logs Útiles

```bash
# Ver logs de Cloud Build
gcloud builds log BUILD_ID

# Ver logs detallados del servicio
gcloud run services logs tail bot-service-staging --region=us-central1

# Ver métricas
gcloud monitoring metrics list --filter="resource.type=cloud_run_revision"
```

## 📞 Soporte

Para problemas o preguntas:
1. Revisar logs del servicio
2. Verificar configuración de secretos
3. Comprobar permisos IAM
4. Consultar documentación de Cloud Run

---

**¡El bot-service está listo para procesar conversaciones inteligentes en la nube!** 🤖☁️