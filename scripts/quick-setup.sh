#!/bin/bash

# Script rápido para configurar el sistema de pruebas
# Ejecuta todo automáticamente sin interacciones

set -e

# Colores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🚀 Configuración rápida del sistema de pruebas${NC}"
echo

# Función para mostrar progreso
show_step() {
    echo -e "${YELLOW}📋 $1${NC}"
}

# Función para mostrar éxito
show_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

# Función para mostrar error
show_error() {
    echo -e "${RED}❌ $1${NC}"
}

# 1. Crear archivo .env.local
show_step "1. Creando archivo .env.local"
cat > .env.local << 'EOF'
# Configuración de desarrollo para IT Bot Service
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=it_bot_service
DB_SSL_MODE=disable
ENVIRONMENT=development
IT_BOT_SERVICE_PORT=8084
LOG_LEVEL=debug
VAULT_ADDR=http://localhost:8200
VAULT_TOKEN=dev-token
VAULT_PATH=secret/microservice
IT_INTEGRATION_SERVICE_URL=http://localhost:8080
EXTERNAL_API_KEY=dev-api-key
EXTERNAL_API_TIMEOUT=30
TEST_DB_HOST=localhost
TEST_DB_PORT=5432
TEST_DB_USER=postgres
TEST_DB_PASSWORD=postgres
TEST_DB_NAME=it_bot_service_test
TEST_DB_SSL_MODE=disable
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0
PROMETHEUS_ENABLED=true
PROMETHEUS_PORT=9090
LOG_FORMAT=json
LOG_OUTPUT=stdout
EOF
show_success "Archivo .env.local creado"

# 2. Verificar Docker
show_step "2. Verificando Docker"
if ! command -v docker &> /dev/null; then
    show_error "Docker no está instalado"
    exit 1
fi

if ! command -v docker-compose &> /dev/null; then
    show_error "Docker Compose no está instalado"
    exit 1
fi
show_success "Docker y Docker Compose disponibles"

# 3. Detener contenedores existentes
show_step "3. Deteniendo contenedores existentes"
docker-compose down 2>/dev/null || true
show_success "Contenedores detenidos"

# 4. Configurar docker-compose con script de pruebas
show_step "4. Configurando Docker Compose"
if ! grep -q "init-test-tables.sql" docker-compose.yml; then
    cp docker-compose.yml docker-compose.yml.backup
    sed -i 's|./scripts/init.sql:/docker-entrypoint-initdb.d/init.sql|./scripts/init.sql:/docker-entrypoint-initdb.d/init.sql\n      - ./scripts/init-test-tables.sql:/docker-entrypoint-initdb.d/init-test-tables.sql|' docker-compose.yml
    show_success "Docker Compose configurado"
else
    show_success "Docker Compose ya configurado"
fi

# 5. Iniciar servicios
show_step "5. Iniciando servicios"
docker-compose up -d postgres vault redis prometheus
show_success "Servicios iniciados"

# 6. Esperar PostgreSQL
show_step "6. Esperando PostgreSQL"
echo "Esperando a que PostgreSQL esté listo..."
until docker-compose exec -T postgres pg_isready -U postgres; do
    echo "PostgreSQL no está listo aún, esperando..."
    sleep 2
done
show_success "PostgreSQL listo"

# 7. Crear base de datos correcta si no existe
show_step "7. Configurando base de datos"
docker-compose exec -T postgres psql -U postgres -c "CREATE DATABASE it_bot_service;" 2>/dev/null || true
docker-compose exec -T postgres psql -U postgres -c "CREATE DATABASE it_bot_service_test;" 2>/dev/null || true
show_success "Bases de datos configuradas"

# 8. Verificar tablas de pruebas
show_step "8. Verificando tablas de pruebas"
sleep 5

if docker-compose exec -T postgres psql -U postgres -d it_bot_service -c "\dt" | grep -q "conditionals"; then
    show_success "Tablas de pruebas creadas"
else
    show_step "Ejecutando script de pruebas manualmente"
    docker-compose exec -T postgres psql -U postgres -d it_bot_service -f /docker-entrypoint-initdb.d/init-test-tables.sql
    show_success "Script de pruebas ejecutado"
fi

# 9. Mostrar resumen
show_step "9. Mostrando resumen de datos"
echo -e "${BLUE}📊 Datos creados:${NC}"
echo "Condicionales:"
docker-compose exec -T postgres psql -U postgres -d it_bot_service -c "SELECT id, name FROM conditionals;" 2>/dev/null || echo "  - No se pudieron mostrar condicionales"

echo "Triggers:"
docker-compose exec -T postgres psql -U postgres -d it_bot_service -c "SELECT id, name FROM triggers;" 2>/dev/null || echo "  - No se pudieron mostrar triggers"

echo "Casos de prueba:"
docker-compose exec -T postgres psql -U postgres -d it_bot_service -c "SELECT id, name FROM test_cases;" 2>/dev/null || echo "  - No se pudieron mostrar casos de prueba"

# 10. Compilar servicio
show_step "10. Compilando servicio"
if command -v go &> /dev/null; then
    go build -o bot-service .
    show_success "Servicio compilado"
else
    show_error "Go no está instalado"
fi

# 11. Información final
show_step "11. Información de conexión"
echo -e "${BLUE}📋 Servicios disponibles:${NC}"
echo "  • PostgreSQL: localhost:5432 (postgres/postgres/it_bot_service)"
echo "  • Vault: http://localhost:8200 (token: dev-token)"
echo "  • Redis: localhost:6379"
echo "  • Prometheus: http://localhost:9090"
echo "  • Bot Service: http://localhost:8084"

echo -e "${BLUE}🔧 Comandos útiles:${NC}"
echo "  • Ejecutar servicio: ./bot-service"
echo "  • Probar APIs: ./scripts/test-api.sh"
echo "  • Ver logs: docker-compose logs"
echo "  • Detener: docker-compose down"

echo
echo -e "${GREEN}🎉 Configuración completada!${NC}"
echo -e "${GREEN}El sistema está listo para usar.${NC}" 