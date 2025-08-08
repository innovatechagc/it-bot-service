#!/bin/bash

# Script para configurar el sistema de pruebas con base de datos externa
# No requiere Docker, solo configuración de entorno

set -e

# Colores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🚀 Configurando sistema de pruebas con base de datos externa${NC}"
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

# Función para mostrar información
show_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

# 1. Crear archivo .env.local
show_step "1. Creando archivo .env.local"
cat > .env.local << 'EOF'
# Configuración de desarrollo para IT Bot Service
# Base de datos PostgreSQL (externa)
DB_HOST=35.227.10.150
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD='p?<MJap]Lqm]LO6G'
DB_NAME=it_bot_service
DB_SSL_MODE=disable

# Configuración del servicio
ENVIRONMENT=development
IT_BOT_SERVICE_PORT=8084
LOG_LEVEL=debug

# Vault (opcional para desarrollo)
VAULT_ADDR=http://localhost:8200
VAULT_TOKEN=dev-token
VAULT_PATH=secret/microservice

# Servicios externos
IT_INTEGRATION_SERVICE_URL=http://localhost:8080
EXTERNAL_API_KEY=dev-api-key
EXTERNAL_API_TIMEOUT=30

# Configuración de pruebas
TEST_DB_HOST=35.227.10.150
TEST_DB_PORT=5432
TEST_DB_USER=postgres
TEST_DB_PASSWORD='p?<MJap]Lqm]LO6G'
TEST_DB_NAME=it_bot_service
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

# 2. Verificar conexión a PostgreSQL
show_step "2. Verificando conexión a PostgreSQL"
if command -v psql &> /dev/null; then
    # Intentar conectar a PostgreSQL usando la configuración externa
    if PGPASSWORD='p?<MJap]Lqm]LO6G' psql -h 35.227.10.150 -U postgres -d postgres -c "SELECT 1;" &> /dev/null; then
        show_success "Conexión a PostgreSQL exitosa"
    else
        show_error "No se pudo conectar a PostgreSQL"
        show_info "Verificando conectividad a 35.227.10.150:5432..."
        
        # Verificar si el puerto está abierto
        if nc -z 35.227.10.150 5432 2>/dev/null; then
            show_info "Puerto 5432 está abierto en 35.227.10.150"
            show_info "Verificando credenciales..."
            
            # Intentar conectar sin especificar base de datos
            if PGPASSWORD='p?<MJap]Lqm]LO6G' psql -h 35.227.10.150 -U postgres -c "SELECT 1;" &> /dev/null; then
                show_success "Conexión exitosa (sin especificar base de datos)"
            else
                show_error "Error de autenticación. Verifica usuario y contraseña."
                exit 1
            fi
        else
            show_error "No se puede conectar al puerto 5432 en 35.227.10.150"
            show_info "Verifica que la IP y puerto sean correctos"
            exit 1
        fi
    fi
else
    show_error "psql no está instalado"
    show_info "Instala PostgreSQL client para continuar"
    exit 1
fi

# 3. Crear bases de datos si no existen
show_step "3. Configurando bases de datos"
PGPASSWORD='p?<MJap]Lqm]LO6G' psql -h 35.227.10.150 -U postgres -c "CREATE DATABASE it_bot_service;" 2>/dev/null || true
show_success "Base de datos configurada"

# 4. Ejecutar script SQL de pruebas
show_step "4. Ejecutando script de pruebas"
if [ -f "./scripts/init-test-tables.sql" ]; then
    PGPASSWORD='p?<MJap]Lqm]LO6G' psql -h 35.227.10.150 -U postgres -d it_bot_service -f ./scripts/init-test-tables.sql
    show_success "Script de pruebas ejecutado"
else
    show_error "Archivo scripts/init-test-tables.sql no encontrado"
    exit 1
fi

# 5. Verificar tablas creadas
show_step "5. Verificando tablas creadas"
echo -e "${BLUE}📊 Tablas creadas:${NC}"
PGPASSWORD='p?<MJap]Lqm]LO6G' psql -h 35.227.10.150 -U postgres -d it_bot_service -c "\dt"

# 6. Mostrar datos de ejemplo
show_step "6. Mostrando datos de ejemplo"
echo -e "${BLUE}📊 Condicionales creados:${NC}"
PGPASSWORD='p?<MJap]Lqm]LO6G' psql -h 35.227.10.150 -U postgres -d it_bot_service -c "SELECT id, name, type FROM conditionals;" 2>/dev/null || echo "  - No se pudieron mostrar condicionales"

echo -e "${BLUE}📊 Triggers creados:${NC}"
PGPASSWORD='p?<MJap]Lqm]LO6G' psql -h 35.227.10.150 -U postgres -d it_bot_service -c "SELECT id, name, event FROM triggers;" 2>/dev/null || echo "  - No se pudieron mostrar triggers"

echo -e "${BLUE}📊 Casos de prueba creados:${NC}"
PGPASSWORD='p?<MJap]Lqm]LO6G' psql -h 35.227.10.150 -U postgres -d it_bot_service -c "SELECT id, name, status FROM test_cases;" 2>/dev/null || echo "  - No se pudieron mostrar casos de prueba"

echo -e "${BLUE}📊 Suites de prueba creadas:${NC}"
PGPASSWORD='p?<MJap]Lqm]LO6G' psql -h 35.227.10.150 -U postgres -d it_bot_service -c "SELECT id, name, status FROM test_suites;" 2>/dev/null || echo "  - No se pudieron mostrar suites de prueba"

# 7. Compilar servicio
show_step "7. Compilando servicio"
if command -v go &> /dev/null; then
    echo "Actualizando dependencias..."
    go mod tidy
    
    echo "Compilando el servicio..."
    if go build -o bot-service .; then
        show_success "Servicio compilado exitosamente"
    else
        show_error "Error al compilar el servicio"
        show_info "Verificando dependencias..."
        go mod download
        go build -o bot-service .
        show_success "Servicio compilado después de descargar dependencias"
    fi
else
    show_error "Go no está instalado"
    show_info "Instala Go para compilar el servicio: https://golang.org/doc/install"
fi

# 8. Información final
show_step "8. Información de conexión"
echo -e "${BLUE}📋 Configuración:${NC}"
echo "  • Base de datos: 35.227.10.150:5432 (postgres/it_bot_service)"
echo "  • Puerto del servicio: 8084"
echo "  • Archivo de configuración: .env.local"

echo -e "${BLUE}🔧 Comandos útiles:${NC}"
echo "  • Ejecutar servicio: ./bot-service"
echo "  • Probar APIs: ./scripts/test-api.sh"
echo "  • Conectar a PostgreSQL: psql -h 35.227.10.150 -U postgres -d it_bot_service"
echo "  • Ver tablas: psql -h 35.227.10.150 -U postgres -d it_bot_service -c \"\\dt\""

# 9. Preguntar si ejecutar el servicio
show_step "9. Ejecutar servicio"
echo -e "${YELLOW}¿Deseas ejecutar el servicio ahora? (y/n)${NC}"
read -r response
if [[ "$response" =~ ^[Yy]$ ]]; then
    show_step "Ejecutando servicio"
    echo "El servicio se ejecutará en http://localhost:8084"
    echo "Presiona Ctrl+C para detener"
    ./bot-service
else
    echo -e "${BLUE}Para ejecutar el servicio manualmente:${NC}"
    echo "  ./bot-service"
    echo "  # o"
    echo "  go run main.go"
fi

# 10. Preguntar si ejecutar pruebas API
show_step "10. Ejecutar pruebas API"
echo -e "${YELLOW}¿Deseas ejecutar las pruebas API ahora? (y/n)${NC}"
read -r response
if [[ "$response" =~ ^[Yy]$ ]]; then
    if [ -f "./scripts/test-api.sh" ]; then
        show_step "Ejecutando pruebas API"
        ./scripts/test-api.sh
    else
        show_error "Script de pruebas API no encontrado"
    fi
else
    echo -e "${BLUE}Para ejecutar las pruebas API manualmente:${NC}"
    echo "  ./scripts/test-api.sh"
fi

echo
echo -e "${GREEN}🎉 Configuración completada!${NC}"
echo -e "${GREEN}El sistema de pruebas con condicionales y triggers está listo para usar.${NC}"
echo
echo -e "${BLUE}📚 Documentación:${NC}"
echo "  • TESTING_FEATURES.md - Documentación completa"
echo "  • scripts/test_conditionals_and_triggers.go - Ejemplos de código"
echo "  • scripts/init-test-tables.sql - Estructura de base de datos" 