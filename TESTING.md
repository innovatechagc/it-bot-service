# 🧪 Bot Service - Guía de Pruebas

Esta guía te ayudará a probar el `it-bot-service` tanto localmente como en la nube.

## 🏠 Pruebas Locales

### 1. Configuración Inicial

```bash
# Clonar y configurar
git clone <repository-url>
cd it-bot-service

# Instalar dependencias
go mod download
```

### 2. Ejecutar el Servicio Localmente

```bash
# Opción 1: Usar el script automatizado (recomendado)
./scripts/run-local.sh

# Opción 2: Ejecutar manualmente
go run .
```

El servicio estará disponible en:
- **API**: http://localhost:8080
- **Health Check**: http://localhost:8080/api/v1/health
- **Swagger**: http://localhost:8080/swagger/index.html

### 3. Crear Datos de Ejemplo

```bash
# Ejecutar script de datos de ejemplo
go run scripts/sample_data.go
```

Esto creará:
- ✅ 1 Bot de ejemplo
- ✅ 1 Flujo conversacional
- ✅ 5 Pasos de diferentes tipos
- ✅ 3 Smart Replies

## 🧪 Pruebas Automatizadas

### Usando curl (Script automatizado)

```bash
# Probar API local
./scripts/test-api.sh local

# Probar API en la nube
./scripts/test-api.sh cloud https://tu-servicio.run.app
```

### Usando Postman

#### 1. Importar Colección

1. Abrir Postman
2. Importar `postman/Bot-Service-API.postman_collection.json`
3. Importar environment:
   - **Local**: `postman/Bot-Service-Local.postman_environment.json`
   - **Cloud**: `postman/Bot-Service-Cloud.postman_environment.json`

#### 2. Configurar Environment

**Para pruebas locales:**
- `base_url`: `http://localhost:8080`

**Para pruebas en la nube:**
- `base_url`: `https://it-bot-service-staging-[HASH]-ue.a.run.app`

#### 3. Ejecutar Pruebas

La colección incluye:

##### 🏥 Health Checks
- Health Check
- Readiness Check

##### 🤖 Bot Management
- Crear bot
- Obtener bots
- Actualizar bot
- Eliminar bot

##### 🔀 Flow Management
- Crear flujo
- Obtener flujos
- Actualizar flujo
- Eliminar flujo

##### 🧩 Step Management
- Crear pasos (mensaje, decisión, IA)
- Actualizar pasos
- Eliminar pasos

##### 🧠 Smart Replies & AI
- Generar respuesta inteligente
- Entrenar intents
- Obtener intents

##### 📨 Message Processing
- Procesar mensajes simples
- Procesar preguntas
- Procesar mensajes de WhatsApp
- Procesar mensajes complejos

## 🌐 Pruebas en la Nube

### 1. Obtener URL del Servicio

```bash
# Obtener URL del servicio desplegado
gcloud run services describe it-bot-service-staging \
    --region=us-east1 \
    --project=innovatech-agc \
    --format='value(status.url)'
```

### 2. Probar Health Checks

```bash
# Health check
curl https://tu-servicio.run.app/api/v1/health

# Readiness check
curl https://tu-servicio.run.app/api/v1/ready
```

### 3. Probar API Completa

```bash
# Usar script automatizado
./scripts/test-api.sh cloud https://tu-servicio.run.app
```

## 📊 Escenarios de Prueba

### Escenario 1: Conversación Básica

1. **Crear Bot**
   ```json
   POST /api/v1/bots
   {
     "name": "Test Bot",
     "owner_id": "owner-001",
     "channel": "web",
     "status": "active"
   }
   ```

2. **Crear Flujo**
   ```json
   POST /api/v1/bots/{bot_id}/flows
   {
     "name": "Welcome Flow",
     "trigger": "hello",
     "is_default": true
   }
   ```

3. **Procesar Mensaje**
   ```json
   POST /api/v1/incoming
   {
     "bot_id": "{bot_id}",
     "user_id": "test-user",
     "content": "hello",
     "channel": "web"
   }
   ```

### Escenario 2: Flujo con IA

1. **Entrenar Intents**
   ```json
   POST /api/v1/bots/{bot_id}/intents/train
   [
     {
       "intent": "greeting",
       "response": "¡Hola! ¿Cómo puedo ayudarte?",
       "confidence": 0.9
     }
   ]
   ```

2. **Generar Smart Reply**
   ```json
   POST /api/v1/bots/{bot_id}/smart-reply
   {
     "prompt": "El usuario pregunta sobre horarios",
     "context": {"language": "es"}
   }
   ```

### Escenario 3: Flujo Multicanal

1. **Mensaje Web**
   ```json
   POST /api/v1/incoming
   {
     "bot_id": "{bot_id}",
     "user_id": "web-user",
     "content": "Hola",
     "channel": "web"
   }
   ```

2. **Mensaje WhatsApp**
   ```json
   POST /api/v1/incoming
   {
     "bot_id": "{bot_id}",
     "user_id": "whatsapp-user",
     "content": "Hola",
     "channel": "whatsapp"
   }
   ```

## 🔍 Debugging y Logs

### Ver Logs Locales

Los logs aparecen en la consola cuando ejecutas el servicio localmente.

### Ver Logs en la Nube

```bash
# Ver logs en tiempo real
gcloud run services logs tail it-bot-service-staging \
    --region=us-east1 \
    --project=innovatech-agc

# Ver logs específicos
gcloud logging read "resource.type=cloud_run_revision AND resource.labels.service_name=it-bot-service-staging" \
    --limit=50 \
    --format=json
```

## 🚨 Troubleshooting

### Problemas Comunes

1. **Servicio no responde**
   - Verificar que esté ejecutándose: `curl http://localhost:8080/api/v1/health`
   - Revisar logs para errores

2. **Error 404 en endpoints**
   - Verificar que la URL base sea correcta
   - Verificar que el servicio esté desplegado

3. **Error 500 en procesamiento**
   - Verificar que existan bots y flujos
   - Revisar logs para errores específicos

4. **Problemas de CORS**
   - El servicio incluye middleware CORS habilitado
   - Verificar headers en requests desde navegador

### Comandos Útiles

```bash
# Verificar estado del servicio
curl -I http://localhost:8080/api/v1/health

# Probar endpoint específico
curl -X POST http://localhost:8080/api/v1/bots \
  -H "Content-Type: application/json" \
  -d '{"name":"Test","owner_id":"test"}'

# Ver métricas
curl http://localhost:8080/metrics
```

## 📈 Métricas y Monitoreo

### Métricas Disponibles

- `http_requests_total` - Total de requests HTTP
- `http_request_duration_seconds` - Duración de requests
- `bot_messages_processed_total` - Mensajes procesados
- `ai_requests_total` - Requests a IA

### Acceder a Métricas

```bash
# Local
curl http://localhost:8080/metrics

# Cloud
curl https://tu-servicio.run.app/metrics
```

## ✅ Checklist de Pruebas

### Funcionalidad Básica
- [ ] Health checks responden correctamente
- [ ] Swagger documentation accesible
- [ ] Crear, leer, actualizar, eliminar bots
- [ ] Crear, leer, actualizar, eliminar flujos
- [ ] Crear, leer, actualizar, eliminar pasos

### Funcionalidad Avanzada
- [ ] Procesamiento de mensajes entrantes
- [ ] Generación de smart replies
- [ ] Entrenamiento de intents
- [ ] Flujos condicionales
- [ ] Sesiones de conversación

### Rendimiento
- [ ] Respuesta < 2 segundos para operaciones simples
- [ ] Respuesta < 5 segundos para operaciones con IA
- [ ] Manejo de múltiples usuarios concurrentes

### Integración
- [ ] Funciona con diferentes canales (web, whatsapp)
- [ ] Manejo correcto de errores
- [ ] Logs informativos

¡El `it-bot-service` está listo para procesar conversaciones inteligentes! 🤖✨