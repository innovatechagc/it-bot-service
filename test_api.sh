#!/bin/bash

echo "🧪 Probando endpoints de testing..."

# Health check
echo "1. Health check:"
curl -s http://localhost:8084/api/v1/health | jq .

# Crear condicional
echo -e "\n2. Crear condicional:"
curl -s -X POST http://localhost:8084/api/v1/conditionals \
  -H "Content-Type: application/json" \
  -d '{"name":"Test Conditional","type":"simple","expression":"user.age > 18","bot_id":"bot-001"}' | jq .

# Crear trigger
echo -e "\n3. Crear trigger:"
curl -s -X POST http://localhost:8084/api/v1/triggers \
  -H "Content-Type: application/json" \
  -d '{"name":"Test Trigger","event":"message_received","action":{"type":"send_response","config":{"message":"Hello!"},"timeout":5000},"bot_id":"bot-001"}' | jq .

# Crear caso de prueba
echo -e "\n4. Crear caso de prueba:"
curl -s -X POST http://localhost:8084/api/v1/test-cases \
  -H "Content-Type: application/json" \
  -d '{"name":"Test Case","bot_id":"bot-001","input":{"message":"Hello","user_id":"user-001"},"expected":{"response":"Hi there!"}}' | jq .

echo -e "\n✅ Pruebas completadas!" 