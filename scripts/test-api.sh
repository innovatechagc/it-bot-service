#!/bin/bash

# Script para probar las APIs de condicionales, triggers y pruebas
# Requiere que el servicio esté ejecutándose en localhost:8084

BASE_URL="http://localhost:8084"
BOT_ID="bot-001"

echo "=== Probando APIs de Condicionales, Triggers y Pruebas ==="
echo "Base URL: $BASE_URL"
echo "Bot ID: $BOT_ID"
echo

# Colores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Función para hacer requests y mostrar resultados
make_request() {
    local method=$1
    local endpoint=$2
    local data=$3
    local description=$4
    
    echo -e "${BLUE}${description}${NC}"
    echo "Endpoint: $method $BASE_URL$endpoint"
    
    if [ -n "$data" ]; then
        echo "Data: $data"
        response=$(curl -s -X $method "$BASE_URL$endpoint" \
            -H "Content-Type: application/json" \
            -d "$data")
    else
        response=$(curl -s -X $method "$BASE_URL$endpoint")
    fi
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓ Success${NC}"
        echo "Response: $response" | jq '.' 2>/dev/null || echo "Response: $response"
    else
        echo -e "${RED}✗ Failed${NC}"
        echo "Error: $response"
    fi
    echo
}

# 1. Crear condicionales
echo -e "${YELLOW}=== 1. Creando Condicionales ===${NC}"

make_request "POST" "/api/v1/conditionals" '{
    "bot_id": "'$BOT_ID'",
    "name": "Usuario Nuevo",
    "description": "Verifica si el usuario es nuevo",
    "expression": "{{user_type}} == \"new\"",
    "type": "simple",
    "priority": 1
}' "Creando condicional: Usuario Nuevo"

make_request "POST" "/api/v1/conditionals" '{
    "bot_id": "'$BOT_ID'",
    "name": "Mensaje de Saludo",
    "description": "Verifica si el mensaje contiene saludos",
    "expression": "{{message}} contains \"hola\" || {{message}} contains \"buenos días\"",
    "type": "complex",
    "priority": 2
}' "Creando condicional: Mensaje de Saludo"

make_request "POST" "/api/v1/conditionals" '{
    "bot_id": "'$BOT_ID'",
    "name": "Email Válido",
    "description": "Verifica si el email tiene formato válido",
    "expression": "{{email}} regex \"^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$\"",
    "type": "regex",
    "priority": 3
}' "Creando condicional: Email Válido"

# 2. Listar condicionales
echo -e "${YELLOW}=== 2. Listando Condicionales ===${NC}"
make_request "GET" "/api/v1/bots/$BOT_ID/conditionals" "" "Listando condicionales del bot"

# 3. Crear triggers
echo -e "${YELLOW}=== 3. Creando Triggers ===${NC}"

make_request "POST" "/api/v1/triggers" '{
    "bot_id": "'$BOT_ID'",
    "name": "Bienvenida Usuario Nuevo",
    "description": "Envía mensaje de bienvenida a usuarios nuevos",
    "event": "message_received",
    "condition": "cond-001",
    "action": {
        "type": "send_message",
        "config": {
            "message": "¡Bienvenido! Soy tu asistente virtual. ¿En qué puedo ayudarte?",
            "channel": "web"
        },
        "timeout": 5000
    },
    "priority": 1,
    "enabled": true
}' "Creando trigger: Bienvenida Usuario Nuevo"

make_request "POST" "/api/v1/triggers" '{
    "bot_id": "'$BOT_ID'",
    "name": "Respuesta a Saludos",
    "description": "Responde automáticamente a saludos",
    "event": "message_received",
    "condition": "cond-002",
    "action": {
        "type": "send_message",
        "config": {
            "message": "¡Hola! ¿Cómo estás? ¿En qué puedo ayudarte hoy?",
            "channel": "web"
        },
        "timeout": 3000
    },
    "priority": 2,
    "enabled": true
}' "Creando trigger: Respuesta a Saludos"

# 4. Listar triggers
echo -e "${YELLOW}=== 4. Listando Triggers ===${NC}"
make_request "GET" "/api/v1/bots/$BOT_ID/triggers" "" "Listando triggers del bot"

# 5. Crear casos de prueba
echo -e "${YELLOW}=== 5. Creando Casos de Prueba ===${NC}"

make_request "POST" "/api/v1/test-cases" '{
    "bot_id": "'$BOT_ID'",
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
        "response": "¡Bienvenido! Soy tu asistente virtual. ¿En qué puedo ayudarte?",
        "conditions": ["cond-001"],
        "triggers": ["trigger-001"]
    },
    "conditions": ["cond-001"],
    "triggers": ["trigger-001"]
}' "Creando caso de prueba: Usuario Nuevo"

make_request "POST" "/api/v1/test-cases" '{
    "bot_id": "'$BOT_ID'",
    "name": "Prueba Saludo",
    "description": "Prueba la respuesta automática a saludos",
    "input": {
        "message": "¡Hola! ¿Cómo estás?",
        "user_id": "user-002",
        "context": {
            "user_type": "existing"
        }
    },
    "expected": {
        "response": "¡Hola! ¿Cómo estás? ¿En qué puedo ayudarte hoy?",
        "conditions": ["cond-002"],
        "triggers": ["trigger-002"]
    },
    "conditions": ["cond-002"],
    "triggers": ["trigger-002"]
}' "Creando caso de prueba: Saludo"

# 6. Listar casos de prueba
echo -e "${YELLOW}=== 6. Listando Casos de Prueba ===${NC}"
make_request "GET" "/api/v1/bots/$BOT_ID/test-cases" "" "Listando casos de prueba del bot"

# 7. Crear suite de pruebas
echo -e "${YELLOW}=== 7. Creando Suite de Pruebas ===${NC}"

make_request "POST" "/api/v1/test-suites" '{
    "bot_id": "'$BOT_ID'",
    "name": "Suite de Pruebas Básicas",
    "description": "Suite de pruebas para funcionalidades básicas del bot",
    "test_cases": ["test-001", "test-002"]
}' "Creando suite de pruebas"

# 8. Listar suites de prueba
echo -e "${YELLOW}=== 8. Listando Suites de Prueba ===${NC}"
make_request "GET" "/api/v1/bots/$BOT_ID/test-suites" "" "Listando suites de prueba del bot"

# 9. Evaluar condicional
echo -e "${YELLOW}=== 9. Evaluando Condicional ===${NC}"
make_request "POST" "/api/v1/conditionals/cond-001/evaluate" '{
    "user_type": "new",
    "first_time": true
}' "Evaluando condicional: Usuario Nuevo"

# 10. Ejecutar trigger
echo -e "${YELLOW}=== 10. Ejecutando Trigger ===${NC}"
make_request "POST" "/api/v1/triggers/trigger-001/execute" '{
    "user_id": "user-001",
    "message": "Hola, soy nuevo aquí",
    "user_type": "new"
}' "Ejecutando trigger: Bienvenida Usuario Nuevo"

# 11. Ejecutar caso de prueba
echo -e "${YELLOW}=== 11. Ejecutando Caso de Prueba ===${NC}"
make_request "POST" "/api/v1/test-cases/test-001/execute" "" "Ejecutando caso de prueba: Usuario Nuevo"

# 12. Ejecutar suite de pruebas
echo -e "${YELLOW}=== 12. Ejecutando Suite de Pruebas ===${NC}"
make_request "POST" "/api/v1/test-suites/suite-001/execute" "" "Ejecutando suite de pruebas"

# 13. Ejecutar múltiples casos de prueba
echo -e "${YELLOW}=== 13. Ejecutando Múltiples Casos de Prueba ===${NC}"
make_request "POST" "/api/v1/test-cases/bulk-execute" '{
    "test_case_ids": ["test-001", "test-002"]
}' "Ejecutando múltiples casos de prueba"

echo -e "${GREEN}=== Pruebas Completadas ===${NC}"
echo "Revisa los logs del servicio para ver detalles de ejecución."
echo "Los datos de ejemplo están disponibles en la base de datos PostgreSQL."