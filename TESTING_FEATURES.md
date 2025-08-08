# Funcionalidades de Pruebas con Condicionales y Triggers

## Resumen

Este documento describe las nuevas funcionalidades implementadas para realizar pruebas usando condicionales y triggers en el sistema de bots conversacionales. Los datos se almacenan en **PostgreSQL** para mantener consistencia transaccional y soporte nativo para triggers de base de datos.

## Arquitectura

### Base de Datos: PostgreSQL

**Decisión**: Se eligió PostgreSQL sobre Firebase NoSQL por las siguientes razones:

1. **Ya configurado** en el proyecto
2. **Mejor para relaciones complejas** entre entidades
3. **Soporte nativo para triggers** de base de datos
4. **Consistencia transaccional** para pruebas
5. **Consultas complejas** para análisis de resultados

### Componentes Principales

1. **Condicionales** (`Conditional`): Expresiones evaluables
2. **Triggers** (`Trigger`): Disparadores de eventos
3. **Casos de Prueba** (`TestCase`): Pruebas individuales
4. **Suites de Prueba** (`TestSuite`): Conjuntos de pruebas

## Entidades

### Conditional

```go
type Conditional struct {
    ID          string                 `json:"id"`
    BotID       string                 `json:"bot_id"`
    Name        string                 `json:"name"`
    Description string                 `json:"description"`
    Expression  string                 `json:"expression"`
    Type        ConditionalType        `json:"type"`
    Priority    int                    `json:"priority"`
    Metadata    map[string]interface{} `json:"metadata,omitempty"`
    CreatedAt   time.Time              `json:"created_at"`
    UpdatedAt   time.Time              `json:"updated_at"`
}
```

**Tipos de Condicionales**:
- `simple`: Condiciones básicas (==, !=)
- `complex`: Condiciones complejas (AND, OR)
- `regex`: Expresiones regulares
- `ai`: Evaluación con IA
- `external`: Condiciones externas

### Trigger

```go
type Trigger struct {
    ID          string                 `json:"id"`
    BotID       string                 `json:"bot_id"`
    Name        string                 `json:"name"`
    Description string                 `json:"description"`
    Event       TriggerEvent           `json:"event"`
    Condition   string                 `json:"condition"` // ID de la condición
    Action      TriggerAction          `json:"action"`
    Priority    int                    `json:"priority"`
    Enabled     bool                   `json:"enabled"`
    Metadata    map[string]interface{} `json:"metadata,omitempty"`
    CreatedAt   time.Time              `json:"created_at"`
    UpdatedAt   time.Time              `json:"updated_at"`
}
```

**Eventos Soportados**:
- `message_received`: Mensaje recibido
- `user_joined`: Usuario se une
- `user_left`: Usuario se va
- `timeout`: Timeout
- `error`: Error
- `custom`: Evento personalizado

### TestCase

```go
type TestCase struct {
    ID          string                 `json:"id"`
    BotID       string                 `json:"bot_id"`
    Name        string                 `json:"name"`
    Description string                 `json:"description"`
    Input       TestInput              `json:"input"`
    Expected    TestExpected           `json:"expected"`
    Conditions  []string               `json:"conditions"` // IDs de condiciones
    Triggers    []string               `json:"triggers"`   // IDs de triggers
    Status      TestStatus             `json:"status"`
    Result      *TestResult            `json:"result,omitempty"`
    Metadata    map[string]interface{} `json:"metadata,omitempty"`
    CreatedAt   time.Time              `json:"created_at"`
    UpdatedAt   time.Time              `json:"updated_at"`
}
```

## API Endpoints

### Condicionales

```
POST   /api/v1/conditionals                    # Crear condicional
GET    /api/v1/conditionals/{id}               # Obtener condicional
PUT    /api/v1/conditionals/{id}               # Actualizar condicional
DELETE /api/v1/conditionals/{id}               # Eliminar condicional
GET    /api/v1/bots/{botId}/conditionals       # Listar condicionales por bot
POST   /api/v1/conditionals/{id}/evaluate      # Evaluar condicional
```

### Triggers

```
POST   /api/v1/triggers                        # Crear trigger
GET    /api/v1/triggers/{id}                   # Obtener trigger
PUT    /api/v1/triggers/{id}                   # Actualizar trigger
DELETE /api/v1/triggers/{id}                   # Eliminar trigger
GET    /api/v1/bots/{botId}/triggers           # Listar triggers por bot
POST   /api/v1/triggers/{id}/execute           # Ejecutar trigger
```

### Casos de Prueba

```
POST   /api/v1/test-cases                      # Crear caso de prueba
GET    /api/v1/test-cases/{id}                 # Obtener caso de prueba
PUT    /api/v1/test-cases/{id}                 # Actualizar caso de prueba
DELETE /api/v1/test-cases/{id}                 # Eliminar caso de prueba
GET    /api/v1/bots/{botId}/test-cases         # Listar casos por bot
POST   /api/v1/test-cases/{id}/execute         # Ejecutar caso de prueba
POST   /api/v1/test-cases/bulk-execute         # Ejecutar múltiples casos
```

### Suites de Prueba

```
POST   /api/v1/test-suites                     # Crear suite de prueba
GET    /api/v1/test-suites/{id}                # Obtener suite de prueba
PUT    /api/v1/test-suites/{id}                # Actualizar suite de prueba
DELETE /api/v1/test-suites/{id}                # Eliminar suite de prueba
GET    /api/v1/bots/{botId}/test-suites        # Listar suites por bot
POST   /api/v1/test-suites/{id}/execute        # Ejecutar suite de prueba
POST   /api/v1/test-suites/{id}/test-cases/{testCaseId}  # Agregar caso a suite
DELETE /api/v1/test-suites/{id}/test-cases/{testCaseId}  # Remover caso de suite
```

## Ejemplos de Uso

### 1. Crear un Condicional

```bash
curl -X POST http://localhost:8080/api/v1/conditionals \
  -H "Content-Type: application/json" \
  -d '{
    "bot_id": "bot-001",
    "name": "Usuario Nuevo",
    "description": "Verifica si el usuario es nuevo",
    "expression": "{{user_type}} == \"new\"",
    "type": "simple",
    "priority": 1
  }'
```

### 2. Crear un Trigger

```bash
curl -X POST http://localhost:8080/api/v1/triggers \
  -H "Content-Type: application/json" \
  -d '{
    "bot_id": "bot-001",
    "name": "Bienvenida Usuario Nuevo",
    "description": "Envía mensaje de bienvenida a usuarios nuevos",
    "event": "message_received",
    "condition": "cond-001",
    "action": {
      "type": "send_message",
      "config": {
        "message": "¡Bienvenido! Soy tu asistente virtual.",
        "channel": "web"
      },
      "timeout": 5000
    },
    "priority": 1,
    "enabled": true
  }'
```

### 3. Crear un Caso de Prueba

```bash
curl -X POST http://localhost:8080/api/v1/test-cases \
  -H "Content-Type: application/json" \
  -d '{
    "bot_id": "bot-001",
    "name": "Prueba Usuario Nuevo",
    "description": "Prueba el flujo de bienvenida para usuarios nuevos",
    "input": {
      "message": "Hola, soy nuevo aquí",
      "user_id": "user-001",
      "context": {
        "user_type": "new",
        "first_time": true
      }
    },
    "expected": {
      "response": "¡Bienvenido! Soy tu asistente virtual.",
      "conditions": ["cond-001"],
      "triggers": ["trigger-001"]
    },
    "conditions": ["cond-001"],
    "triggers": ["trigger-001"]
  }'
```

### 4. Ejecutar un Caso de Prueba

```bash
curl -X POST http://localhost:8080/api/v1/test-cases/test-001/execute
```

### 5. Crear una Suite de Pruebas

```bash
curl -X POST http://localhost:8080/api/v1/test-suites \
  -H "Content-Type: application/json" \
  -d '{
    "bot_id": "bot-001",
    "name": "Suite de Pruebas Básicas",
    "description": "Suite de pruebas para funcionalidades básicas",
    "test_cases": ["test-001", "test-002", "test-003", "test-004"]
  }'
```

### 6. Ejecutar una Suite de Pruebas

```bash
curl -X POST http://localhost:8080/api/v1/test-suites/suite-001/execute
```

## Expresiones de Condicionales

### Sintaxis Soportada

1. **Comparaciones simples**:
   ```
   {{variable}} == "valor"
   {{variable}} != "valor"
   ```

2. **Contenido**:
   ```
   {{texto}} contains "palabra"
   ```

3. **Expresiones regulares**:
   ```
   {{email}} regex "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$"
   ```

4. **Operadores lógicos**:
   ```
   {{cond1}} && {{cond2}}
   {{cond1}} || {{cond2}}
   ```

5. **Variables de contexto**:
   ```
   {{user_type}} == "new"
   {{subscription_type}} == "premium" && {{subscription_active}} == true
   ```

## Configuración de Base de Datos

### 1. Ejecutar el script de inicialización

```bash
# Para desarrollo
psql -h localhost -U postgres -d microservice_dev -f scripts/init-test-tables.sql

# Para testing
psql -h localhost -U postgres -d microservice_test -f scripts/init-test-tables.sql
```

### 2. Verificar las tablas creadas

```sql
-- Listar tablas
\dt

-- Verificar datos de ejemplo
SELECT * FROM conditionals;
SELECT * FROM triggers;
SELECT * FROM test_cases;
SELECT * FROM test_suites;
```

## Servicios Implementados

### ConditionalService
- Gestión de condicionales
- Evaluación de expresiones
- Soporte para diferentes tipos de condiciones

### TriggerService
- Gestión de triggers
- Ejecución de acciones
- Procesamiento de eventos

### TestService
- Gestión de casos de prueba
- Ejecución de pruebas
- Verificación de resultados

### TestSuiteService
- Gestión de suites de prueba
- Ejecución en lote
- Análisis de resultados

## Estados de Pruebas

### TestStatus
- `pending`: Pendiente de ejecución
- `running`: En ejecución
- `passed`: Exitoso
- `failed`: Fallido
- `skipped`: Omitido

### TestSuiteStatus
- `pending`: Pendiente de ejecución
- `running`: En ejecución
- `passed`: Todos los casos exitosos
- `failed`: Todos los casos fallidos
- `partial`: Algunos casos exitosos, otros fallidos

## Monitoreo y Logs

### Logs de Ejecución

Los servicios registran logs detallados para:
- Evaluación de condicionales
- Ejecución de triggers
- Resultados de pruebas
- Errores y excepciones

### Métricas Disponibles

- Tiempo de ejecución por caso de prueba
- Tasa de éxito por suite
- Número de condiciones evaluadas
- Número de triggers ejecutados

## Próximos Pasos

1. **Implementar repositorios PostgreSQL** para las nuevas entidades
2. **Integrar con el sistema de bots** existente
3. **Agregar validaciones** adicionales
4. **Implementar más tipos de condiciones** (IA, externas)
5. **Crear interfaz web** para gestión de pruebas
6. **Agregar reportes** y dashboards
7. **Implementar notificaciones** de resultados

## Archivos Creados

- `internal/domain/entities.go` (actualizado)
- `internal/domain/repositories.go` (actualizado)
- `internal/services/conditional_service.go` (nuevo)
- `internal/services/test_service.go` (nuevo)
- `internal/handlers/test_handlers.go` (nuevo)
- `scripts/test_conditionals_and_triggers.go` (nuevo)
- `scripts/init-test-tables.sql` (nuevo)
- `TESTING_FEATURES.md` (nuevo)

## Consideraciones de Rendimiento

1. **Índices de base de datos** optimizados para consultas frecuentes
2. **Evaluación lazy** de condicionales
3. **Ejecución asíncrona** de triggers
4. **Caché de resultados** de evaluaciones
5. **Timeouts configurables** para evitar bloqueos

## Seguridad

1. **Validación de expresiones** para prevenir inyección
2. **Sanitización de inputs** en evaluaciones
3. **Límites de tiempo** en ejecuciones
4. **Logs de auditoría** para todas las operaciones
5. **Control de acceso** basado en roles 