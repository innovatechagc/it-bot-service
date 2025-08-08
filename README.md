# 🤖 it-bot-service - Orquestador de Agentes y Flujos Conversacionales

Microservicio avanzado de orquestación que combina **gestión de flujos tipo n8n** con **coordinación de agentes MCP** para crear experiencias conversacionales inteligentes y automatizadas. 

## 🎯 **Propósito Principal**

El `it-bot-service` actúa como el **cerebro orquestador** que:

- 🔀 **Gestiona flujos** de automatización visual tipo n8n
- 🤖 **Instancia y coordina MCPs** según flujo, canal o usuario  
- 🧠 **Mantiene contexto** y memoria de conversaciones
- 🔗 **Conecta servicios** internos y externos
- 📨 **Procesa mensajes** multicanal con respuestas inteligentes

### **Flujo de Orquestación:**
```
Mensaje → Flujo → Agente MCP → Contexto → Respuesta → Integración
```

El servicio **instancia MCPs** cuando el flujo lo requiere, les **pasa contexto y objetivos**, recibe la **salida del agente** y actúa en consecuencia (responde, llama a otro microservicio, etc.), llevando **registro completo** de conversaciones, pasos y decisiones.

## 🚀 Características Implementadas

### ✅ **Core Funcional (Implementado)**
- **🔀 Gestión de Flujos**: CRUD completo de flujos tipo n8n (crear, leer, actualizar, eliminar)
- **🧩 Pasos Modulares**: 5 tipos de pasos (message, decision, input, api_call, ai)
- **📨 Procesamiento Multicanal**: Web, WhatsApp, Telegram, Slack
- **🧠 Context Manager**: Memoria corta con sesiones y variables de flujo
- **🤖 Smart Replies**: Respuestas inteligentes basadas en IA e intents
- **⚡ Ejecución de Flujos**: Motor de ejecución condicional
- **🔗 API REST Completa**: Todos los endpoints para integración

### ⚠️ **En Desarrollo (Parcial)**
- **🎯 Orquestación MCP**: Estructura básica, falta instanciación de agentes
- **🔌 Integraciones**: Tipo api_call implementado, falta ejecución real
- **📊 Eventos**: Estructura en metadata, falta procesamiento completo
- **🎨 UI Visual**: API lista, interfaz visual marcada para futuro

### ❌ **Pendiente (Crítico)**
- **🤖 Sistema de Agentes MCP**: Instanciación y coordinación de MCPs
- **🧠 Memoria Persistente**: Almacenamiento a largo plazo
- **🔧 Adaptadores/Plugins**: Sistema extensible de conectores
- **📡 gRPC**: Soporte para llamadas gRPC

## 📊 Estado de Funcionalidades Requeridas

| Funcionalidad | Estado | Completitud |
|---------------|--------|-------------|
| 🔀 Gestión de flujos tipo n8n | ✅ **Implementado** | 85% |
| 🤖 Orquestación de agentes MCP | ❌ **Pendiente** | 10% |
| 📥📤 Entradas y salidas | ⚠️ **Parcial** | 70% |
| 🧠 Context Manager | ⚠️ **Parcial** | 60% |
| 🔧 Interfaz modular | ⚠️ **Parcial** | 65% |
| 🔗 Interoperabilidad | ❌ **Pendiente** | 20% |
| 🎨 UI para flujos | ❌ **Futuro** | 0% |

**Estado General**: 🟡 **Funcional Básico** (50% completitud) - Listo para casos de uso simples, requiere desarrollo MCP para funcionalidad completa.

## 📁 Estructura del Proyecto

```
├── internal/                    # Código interno del bot-service
│   ├── ai/                     # Cliente de IA (OpenAI, Vertex AI)
│   ├── config/                 # Configuración del servicio
│   ├── domain/                 # Entidades y repositorios
│   │   ├── entities.go         # Bot, BotFlow, BotStep, SmartReply
│   │   └── repositories.go     # Interfaces de persistencia
│   ├── handlers/               # Handlers HTTP del bot
│   │   ├── handlers.go         # Health checks
│   │   └── bot_handlers.go     # CRUD de bots, flujos, pasos
│   ├── middleware/             # Middleware personalizado
│   ├── repositories/           # Implementaciones mock
│   ├── services/               # Lógica de negocio
│   │   ├── bot.go             # Orquestación principal
│   │   ├── bot_flow.go        # Gestión de flujos
│   │   ├── bot_step.go        # Gestión de pasos
│   │   ├── smart_reply.go     # IA y respuestas inteligentes
│   │   └── conversation.go    # Manejo de sesiones
│   └── testing/               # Utilidades de testing
├── pkg/                       # Paquetes reutilizables
│   ├── logger/               # Logger estructurado
│   ├── vault/                # Cliente de Vault
│   ├── events/               # Sistema de eventos
│   ├── featureflags/         # Feature flags
│   └── tracing/              # Tracing distribuido
├── postman/                  # Colecciones de Postman
│   ├── Bot-Service-API.postman_collection.json
│   ├── Bot-Service-Local.postman_environment.json
│   └── Bot-Service-Cloud.postman_environment.json
├── scripts/                  # Scripts de utilidad
│   ├── run-local.sh         # Ejecutar localmente
│   ├── test-api.sh          # Pruebas automatizadas
│   ├── sample_data.go       # Datos de ejemplo
│   ├── deploy.sh            # Script de deployment
│   └── setup-gcp.sh         # Configuración de GCP
├── deploy/                   # Configuraciones de deployment
│   ├── cloudrun-staging.yaml
│   └── cloudrun-production.yaml
├── monitoring/               # Configuración de monitoreo
├── tests/                    # Tests de integración y e2e
├── cloudbuild.yaml          # Configuración de Cloud Build
├── Dockerfile               # Imagen optimizada para Cloud Run
├── FUNCTIONALITY_ANALYSIS.md # Análisis detallado de funcionalidades
├── TESTING.md               # Guía completa de pruebas
└── DEPLOYMENT.md            # Guía de deployment en GCP
```

## 🛠️ Configuración Inicial

### 1. Clonar y configurar el proyecto

```bash
# Clonar el repositorio
git clone <repository-url>
cd it-bot-service

# Instalar dependencias
go mod download
make deps
```

### 2. Configurar variables de entorno

El archivo `.env.local` ya está configurado para desarrollo:

```bash
# Configuración del Bot Service
ENVIRONMENT=development
PORT=8080
LOG_LEVEL=debug

# APIs de IA (usar claves reales para funcionalidad completa)
OPENAI_API_KEY=sk-test-key-for-local-development
VERTEX_AI_PROJECT=innovatech-agc
VERTEX_AI_LOCATION=us-east1

# Base de datos (opcional para desarrollo)
DB_HOST=localhost
DB_PORT=5432
DB_USER=bot_service_user
DB_PASSWORD=local_password
DB_NAME=bot_service_dev

# Redis para sesiones (opcional)
REDIS_HOST=localhost
REDIS_PORT=6379

# URLs de servicios relacionados
MESSAGING_SERVICE_URL=http://localhost:8081
USER_SERVICE_URL=http://localhost:8082
```

## 🚀 Desarrollo Local

### Opción 1: Script Automatizado (Recomendado)

```bash
# Ejecutar con datos de ejemplo incluidos
./scripts/run-local.sh
```

### Opción 2: Ejecutar directamente

```bash
# Crear datos de ejemplo
go run scripts/sample_data.go

# Compilar y ejecutar
make build
make run

# O directamente
go run .
```

### Opción 3: Con Docker Compose

```bash
# Levantar entorno completo (opcional)
make docker-dev

# Detener servicios
make docker-down
```

**Servicios disponibles:**
- **🤖 Bot Service API**: http://localhost:8080
- **📚 Swagger Documentation**: http://localhost:8080/swagger/index.html
- **🏥 Health Check**: http://localhost:8080/api/v1/health
- **📊 Metrics**: http://localhost:8080/metrics

## 🧪 Testing

```bash
# Ejecutar tests
make test

# Tests con cobertura
make test-coverage

# Tests con Docker
make docker-test

# Linting
make lint
```

## 📊 API Endpoints del Bot Service

### Health Checks
- `GET /api/v1/health` - Estado del servicio
- `GET /api/v1/ready` - Readiness check

### 🤖 Gestión de Bots
- `GET /api/v1/bots` - Lista bots por usuario o tenant
- `GET /api/v1/bots/:id` - Detalle de un bot específico
- `POST /api/v1/bots` - Crear nuevo bot
- `PATCH /api/v1/bots/:id` - Editar bot existente
- `DELETE /api/v1/bots/:id` - Eliminar o desactivar bot

### 🔀 Gestión de Flujos
- `GET /api/v1/bots/:id/flows` - Lista flujos del bot
- `POST /api/v1/bots/:id/flows` - Crear flujo conversacional
- `GET /api/v1/flows/:id` - Obtener un flujo con sus pasos
- `PATCH /api/v1/flows/:id` - Editar un flujo
- `DELETE /api/v1/flows/:id` - Eliminar un flujo

### 🧩 Gestión de Pasos
- `POST /api/v1/flows/:id/steps` - Agregar paso a un flujo
- `PATCH /api/v1/steps/:id` - Editar paso
- `DELETE /api/v1/steps/:id` - Eliminar paso

### 🧠 IA / Smart Replies
- `POST /api/v1/bots/:id/smart-reply` - Consulta rápida a IA (prompt + contexto)
- `POST /api/v1/bots/:id/intents/train` - Entrenar respuestas automáticas
- `GET /api/v1/bots/:id/intents` - Listar intents configurados

### 📨 Procesamiento de Mensajes
- `POST /api/v1/incoming` - Recibe mensaje entrante desde messaging-service y responde según flujo

### Métricas y Documentación
- `GET /metrics` - Métricas de Prometheus
- `GET /swagger/index.html` - Documentación Swagger completa

## 🔧 Configuración por Entornos

### Desarrollo Local
- Archivo: `.env.local`
- Base de datos: PostgreSQL local
- Vault: Opcional (comentado por defecto)
- Logs: Debug level

### Testing/QA
- Archivo: `.env.test`
- Base de datos: PostgreSQL de testing
- Vault: Instancia de testing
- Logs: Info level

### Producción
- Archivo: `.env.production`
- Variables desde GCP Secret Manager o Vault
- SSL requerido para BD
- Logs: Warn level

## 🐳 Docker

### Desarrollo
```bash
# Construir imagen
make docker-build

# Ejecutar contenedor
make docker-run
```

### Testing
```bash
# Ejecutar tests en contenedor
make docker-test
```

## ☁️ Despliegue en GCP Cloud Run

### Setup Inicial (Solo una vez)
```bash
# Configurar recursos de GCP automáticamente
./scripts/setup-gcp.sh innovatech-agc us-east1
```

### Deploy a Staging
```bash
# Usando script automatizado (recomendado)
make deploy-staging

# O directamente con Cloud Build
gcloud builds submit --config cloudbuild.yaml \
    --substitutions _ENVIRONMENT=staging,_REGION=us-east1 \
    --project=innovatech-agc
```

### Deploy a Producción
```bash
# Usando script automatizado
make deploy-prod

# O directamente con Cloud Build
gcloud builds submit --config cloudbuild.yaml \
    --substitutions _ENVIRONMENT=production,_REGION=us-east1 \
    --project=innovatech-agc
```

**Servicios desplegados:**
- **Staging**: `it-bot-service-staging` en Cloud Run
- **Production**: `it-bot-service-production` en Cloud Run
- **Imágenes**: Almacenadas en `gcr.io/innovatech-agc/it-bot-service`

## 🔐 Manejo de Secretos

### Con Vault (Recomendado)
```go
// Ejemplo de uso
vaultClient, err := vault.NewClient(cfg.VaultConfig)
secrets, err := vaultClient.GetSecret("secret/myapp/database")
password := secrets["password"].(string)
```

### Variables de Entorno
Para desarrollo local, usar archivos `.env.*`

## 📈 Monitoreo y Métricas

### Métricas Disponibles
- `http_requests_total` - Total de requests HTTP
- `http_request_duration_seconds` - Duración de requests

### Prometheus
Configuración en `monitoring/prometheus.yml`

## 🧪 Pruebas Completas

### Pruebas Automatizadas con curl
```bash
# Probar API local
./scripts/test-api.sh local

# Probar API en la nube
./scripts/test-api.sh cloud https://tu-servicio.run.app
```

### Pruebas con Postman
1. **Importar colección**: `postman/Bot-Service-API.postman_collection.json`
2. **Importar environment**: 
   - Local: `postman/Bot-Service-Local.postman_environment.json`
   - Cloud: `postman/Bot-Service-Cloud.postman_environment.json`
3. **Ejecutar pruebas** de todos los endpoints

### Escenarios de Prueba Incluidos
- ✅ **Health Checks** - Verificar estado del servicio
- ✅ **Bot Management** - CRUD completo de bots
- ✅ **Flow Management** - Gestión de flujos conversacionales
- ✅ **Step Management** - Pasos de diferentes tipos
- ✅ **Smart Replies** - Respuestas inteligentes con IA
- ✅ **Message Processing** - Procesamiento multicanal
- ✅ **Load Testing** - Pruebas de carga

## 🎯 Próximos Desarrollos (Roadmap)

### 🔥 **Fase 1: Integración MCP (Crítico)**
- [ ] Implementar `MCPOrchestrator` para instanciar agentes
- [ ] Sistema de coordinación de múltiples MCPs
- [ ] Paso de contexto entre agentes
- [ ] Manejo de estado de agentes

### ⚡ **Fase 2: Integraciones Robustas**
- [ ] Llamadas HTTP/gRPC reales en pasos `api_call`
- [ ] Sistema de adaptadores extensible
- [ ] Conectores para servicios externos
- [ ] Manejo avanzado de errores y reintentos

### 📚 **Fase 3: Memoria Persistente**
- [ ] Base de datos para memoria a largo plazo
- [ ] Indexación y búsqueda de contexto histórico
- [ ] Analytics de conversaciones
- [ ] Métricas de rendimiento de agentes

### 🎨 **Fase 4: UI Visual (Futuro)**
- [ ] Panel web para gestión de flujos
- [ ] Editor drag-and-drop tipo n8n
- [ ] Monitoreo en tiempo real
- [ ] Dashboard de analytics

## 📝 Comandos Útiles

```bash
# Ver todos los comandos disponibles
make help

# Desarrollo
make deps          # Instalar dependencias
make build         # Compilar
make run           # Ejecutar
make test          # Tests
make lint          # Linting
make format        # Formatear código

# Docker
make docker-build  # Construir imagen
make docker-dev    # Entorno completo
make docker-test   # Tests en Docker

# Documentación
make swagger       # Generar docs Swagger
```

## 🤝 Contribución

1. Fork el proyecto
2. Crear feature branch (`git checkout -b feature/nueva-funcionalidad`)
3. Commit cambios (`git commit -am 'Agregar nueva funcionalidad'`)
4. Push al branch (`git push origin feature/nueva-funcionalidad`)
5. Crear Pull Request

## 📄 Licencia

Este proyecto está bajo la Licencia MIT - ver el archivo [LICENSE](LICENSE) para detalles.

## 🆘 Soporte

Para preguntas o problemas:
1. Revisar la documentación
2. Buscar en issues existentes
3. Crear nuevo issue con detalles del problema

---

**Nota**: Este template incluye ejemplos comentados para facilitar el desarrollo. Descomenta y configura según las necesidades de tu microservicio.