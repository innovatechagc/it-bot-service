#!/bin/bash

echo "🧪 Probando endpoints de testing (versión extendida)..."

# Health check
echo "1. Health check:"
curl -s http://localhost:8084/api/v1/health | jq .

# Crear condicional
echo -e "\n2. Crear condicional:"
CONDITIONAL_RESPONSE=$(curl -s -X POST http://localhost:8084/api/v1/conditionals \
  -H "Content-Type: application/json" \
  -d '{"name":"Test Conditional","type":"simple","expression":"user.age > 18","bot_id":"bot-001"}')
echo $CONDITIONAL_RESPONSE | jq .
CONDITIONAL_ID=$(echo $CONDITIONAL_RESPONSE | jq -r '.data.id')

# Obtener condicional por ID
echo -e "\n3. Obtener condicional por ID:"
curl -s http://localhost:8084/api/v1/conditionals/$CONDITIONAL_ID | jq .

# Crear trigger
echo -e "\n4. Crear trigger:"
TRIGGER_RESPONSE=$(curl -s -X POST http://localhost:8084/api/v1/triggers \
  -H "Content-Type: application/json" \
  -d '{"name":"Test Trigger","event":"message_received","action":{"type":"send_response","config":{"message":"Hello!"},"timeout":5000},"bot_id":"bot-001"}')
echo $TRIGGER_RESPONSE | jq .
TRIGGER_ID=$(echo $TRIGGER_RESPONSE | jq -r '.data.id')

# Obtener trigger por ID
echo -e "\n5. Obtener trigger por ID:"
curl -s http://localhost:8084/api/v1/triggers/$TRIGGER_ID | jq .

# Crear caso de prueba
echo -e "\n6. Crear caso de prueba:"
TEST_CASE_RESPONSE=$(curl -s -X POST http://localhost:8084/api/v1/test-cases \
  -H "Content-Type: application/json" \
  -d '{"name":"Test Case","bot_id":"bot-001","input":{"message":"Hello","user_id":"user-001"},"expected":{"response":"Hi there!"}}')
echo $TEST_CASE_RESPONSE | jq .
TEST_CASE_ID=$(echo $TEST_CASE_RESPONSE | jq -r '.data.id')

# Obtener caso de prueba por ID
echo -e "\n7. Obtener caso de prueba por ID:"
curl -s http://localhost:8084/api/v1/test-cases/$TEST_CASE_ID | jq .

# Crear suite de prueba
echo -e "\n8. Crear suite de prueba:"
TEST_SUITE_RESPONSE=$(curl -s -X POST http://localhost:8084/api/v1/test-suites \
  -H "Content-Type: application/json" \
  -d '{"name":"Test Suite","bot_id":"bot-001","description":"Suite de pruebas básicas"}')
echo $TEST_SUITE_RESPONSE | jq .
TEST_SUITE_ID=$(echo $TEST_SUITE_RESPONSE | jq -r '.data.id')

# Obtener suite de prueba por ID
echo -e "\n9. Obtener suite de prueba por ID:"
curl -s http://localhost:8084/api/v1/test-suites/$TEST_SUITE_ID | jq .

# Agregar caso de prueba al suite
echo -e "\n10. Agregar caso de prueba al suite:"
curl -s -X POST http://localhost:8084/api/v1/test-suites/$TEST_SUITE_ID/test-cases \
  -H "Content-Type: application/json" \
  -d "{\"test_case_id\":\"$TEST_CASE_ID\"}" | jq .

# Listar condicionales por bot
echo -e "\n11. Listar condicionales por bot:"
curl -s http://localhost:8084/api/v1/conditionals/bot/bot-001 | jq .

# Listar triggers por bot
echo -e "\n12. Listar triggers por bot:"
curl -s http://localhost:8084/api/v1/triggers/bot/bot-001 | jq .

# Listar casos de prueba por bot
echo -e "\n13. Listar casos de prueba por bot:"
curl -s http://localhost:8084/api/v1/test-cases/bot/bot-001 | jq .

# Listar suites de prueba por bot
echo -e "\n14. Listar suites de prueba por bot:"
curl -s http://localhost:8084/api/v1/test-suites/bot/bot-001 | jq .

echo -e "\n✅ Pruebas extendidas completadas!" 