#!/bin/bash

# Script completo para configurar y ejecutar el sistema de pruebas
# Incluye: configuración de entorno, base de datos, y pruebas

set -e

# Colores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🚀 Configurando y ejecutando sistema de pruebas con condicionales y triggers${NC}"
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

# 1. Crear archivo .env.local si no existe
show_step "1. Configurando archivo .env.local"
if [ ! -f .env.local ]; then
    cat > .env.local << 'EOF'
# Configuración de desarrollo para IT Bot Service
# Base de datos PostgreSQL
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=it_bot_service
DB_SSL_MODE=disable

# Configuración del servicio
ENVIRONMENT=development
IT_BOT_SERVICE_PORT=8084
LOG_LEVEL=debug

# Vault
VAULT_ADDR=http://localhost:8200
VAULT_TOKEN=dev-token
VAULT_PATH=secret/microservice

# Servicios externos
IT_INTEGRATION_SERVICE_URL=http://localhost:8080
EXTERNAL_API_KEY=dev-api-key
EXTERNAL_API_TIMEOUT=30

# Configuración de pruebas
TEST_DB_HOST=localhost
TEST_DB_PORT=5432
TEST_DB_USER=postgres
TEST_DB_PASSWORD=postgres
TEST_DB_NAME=it_bot_service_test
TEST_DB_SSL_MODE=disable

# Configuración de Redis (opcional)
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# Configuración de Prometheus
PROMETHEUS_ENABLED=true
PROMETHEUS_PORT=9090

# Configuración de logs
LOG_FORMAT=json
LOG_OUTPUT=stdout
EOF
    show_success "Archivo .env.local creado"
else
    show_success "Archivo .env.local ya existe"
fi

# 2. Verificar Docker y Docker Compose
show_step "2. Verificando Docker y Docker Compose"
if ! command -v docker &> /dev/null; then
    show_error "Docker no está instalado"
    exit 1
fi

if ! command -v docker-compose &> /dev/null; then
    show_error "Docker Compose no está instalado"
    exit 1
fi

show_success "Docker y Docker Compose están disponibles"

# 3. Detener contenedores existentes si están corriendo
show_step "3. Deteniendo contenedores existentes"
docker-compose down 2>/dev/null || true
show_success "Contenedores detenidos"

# 4. Modificar docker-compose.yml para incluir el script de pruebas
show_step "4. Configurando Docker Compose con script de pruebas"
if ! grep -q "init-test-tables.sql" docker-compose.yml; then
    # Crear un docker-compose temporal con el script de pruebas
    cp docker-compose.yml docker-compose.yml.backup
    
    # Agregar el script de pruebas al volumen de postgres
    sed -i 's|./scripts/init.sql:/docker-entrypoint-initdb.d/init.sql|./scripts/init.sql:/docker-entrypoint-initdb.d/init.sql\n      - ./scripts/init-test-tables.sql:/docker-entrypoint-initdb.d/init-test-tables.sql|' docker-compose.yml
    
    show_success "Docker Compose configurado con script de pruebas"
else
    show_success "Docker Compose ya está configurado"
fi

# 5. Iniciar servicios
show_step "5. Iniciando servicios (PostgreSQL, Vault, Redis, Prometheus)"
docker-compose up -d postgres vault redis prometheus

# 6. Esperar a que PostgreSQL esté listo
show_step "6. Esperando a que PostgreSQL esté listo"
echo "Esperando a que PostgreSQL esté disponible..."
until docker-compose exec -T postgres pg_isready -U postgres; do
    echo "PostgreSQL no está listo aún, esperando..."
    sleep 2
done
show_success "PostgreSQL está listo"

# 7. Crear base de datos correcta si no existe
show_step "7. Configurando base de datos"
docker-compose exec -T postgres psql -U postgres -c "CREATE DATABASE it_bot_service;" 2>/dev/null || true
docker-compose exec -T postgres psql -U postgres -c "CREATE DATABASE it_bot_service_test;" 2>/dev/null || true
show_success "Bases de datos configuradas"

# 8. Verificar que las tablas de pruebas se crearon
show_step "8. Verificando tablas de pruebas"
sleep 5  # Dar tiempo para que se ejecuten los scripts

# Verificar si las tablas existen
if docker-compose exec -T postgres psql -U postgres -d it_bot_service -c "\dt" | grep -q "conditionals"; then
    show_success "Tablas de pruebas creadas correctamente"
else
    show_error "Las tablas de pruebas no se crearon. Ejecutando script manualmente..."
    
    # Ejecutar script manualmente
    docker-compose exec -T postgres psql -U postgres -d it_bot_service -f /docker-entrypoint-initdb.d/init-test-tables.sql || {
        show_error "Error ejecutando script de pruebas"
        exit 1
    }
    show_success "Script de pruebas ejecutado manualmente"
fi

# 9. Mostrar datos de ejemplo
show_step "9. Mostrando datos de ejemplo"
echo -e "${BLUE}📊 Condicionales creados:${NC}"
docker-compose exec -T postgres psql -U postgres -d it_bot_service -c "SELECT id, name, type FROM conditionals;"

echo -e "${BLUE}📊 Triggers creados:${NC}"
docker-compose exec -T postgres psql -U postgres -d it_bot_service -c "SELECT id, name, event FROM triggers;"

echo -e "${BLUE}📊 Casos de prueba creados:${NC}"
docker-compose exec -T postgres psql -U postgres -d it_bot_service -c "SELECT id, name, status FROM test_cases;"

echo -e "${BLUE}📊 Suites de prueba creadas:${NC}"
docker-compose exec -T postgres psql -U postgres -d it_bot_service -c "SELECT id, name, status FROM test_suites;"

# 10. Compilar y ejecutar el servicio (opcional)
show_step "10. Compilando el servicio"
if command -v go &> /dev/null; then
    echo "Compilando el servicio..."
    go build -o bot-service .
    show_success "Servicio compilado"
    
    # Preguntar si quiere ejecutar el servicio
    echo -e "${YELLOW}¿Deseas ejecutar el servicio ahora? (y/n)${NC}"
    read -r response
    if [[ "$response" =~ ^[Yy]$ ]]; then
        show_step "11. Ejecutando el servicio"
        echo "El servicio se ejecutará en http://localhost:8084"
        echo "Presiona Ctrl+C para detener"
        ./bot-service
    else
        echo -e "${BLUE}Para ejecutar el servicio manualmente:${NC}"
        echo "  ./bot-service"
        echo "  # o"
        echo "  go run main.go"
    fi
else
    show_error "Go no está instalado. Instala Go para compilar el servicio."
    echo -e "${BLUE}Para instalar Go:${NC}"
    echo "  https://golang.org/doc/install"
fi

# 11. Mostrar información de conexión
show_step "12. Información de conexión"
echo -e "${BLUE}📋 Servicios disponibles:${NC}"
echo "  • PostgreSQL: localhost:5432 (user: postgres, pass: postgres, db: it_bot_service)"
echo "  • Vault: http://localhost:8200 (token: dev-token)"
echo "  • Redis: localhost:6379"
echo "  • Prometheus: http://localhost:9090"
echo "  • Bot Service: http://localhost:8084 (cuando se ejecute)"

# 12. Mostrar comandos útiles
show_step "13. Comandos útiles"
echo -e "${BLUE}🔧 Comandos útiles:${NC}"
echo "  • Ver logs de PostgreSQL: docker-compose logs postgres"
echo "  • Conectar a PostgreSQL: docker-compose exec postgres psql -U postgres -d it_bot_service"
echo "  • Ejecutar pruebas API: ./scripts/test-api.sh"
echo "  • Detener servicios: docker-compose down"
echo "  • Reiniciar servicios: docker-compose restart"

# 13. Ejecutar script de pruebas API (opcional)
show_step "14. Ejecutando script de pruebas API"
echo -e "${YELLOW}¿Deseas ejecutar las pruebas API ahora? (y/n)${NC}"
read -r response
if [[ "$response" =~ ^[Yy]$ ]]; then
    if [ -f "./scripts/test-api.sh" ]; then
        echo "Ejecutando pruebas API..."
        ./scripts/test-api.sh
    else
        show_error "Script de pruebas API no encontrado"
    fi
else
    echo -e "${BLUE}Para ejecutar las pruebas API manualmente:${NC}"
    echo "  ./scripts/test-api.sh"
fi

echo
echo -e "${GREEN}🎉 Configuración completada exitosamente!${NC}"
echo -e "${GREEN}El sistema de pruebas con condicionales y triggers está listo para usar.${NC}"
echo
echo -e "${BLUE}📚 Documentación:${NC}"
echo "  • TESTING_FEATURES.md - Documentación completa"
echo "  • scripts/test_conditionals_and_triggers.go - Ejemplos de código"
echo "  • scripts/init-test-tables.sql - Estructura de base de datos" 