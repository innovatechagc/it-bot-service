#!/bin/bash

# Script para ejecutar el bot-service localmente con datos de ejemplo

set -e

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🤖 Starting Bot Service locally...${NC}"

# Verificar que go esté instalado
if ! command -v go &> /dev/null; then
    echo -e "${RED}Error: Go is not installed${NC}"
    exit 1
fi

# Cargar variables de entorno
if [ -f .env.local ]; then
    echo -e "${YELLOW}📋 Loading local environment variables...${NC}"
    export $(cat .env.local | grep -v '^#' | xargs)
else
    echo -e "${YELLOW}⚠️  .env.local not found, using defaults${NC}"
fi

# Compilar la aplicación
echo -e "${YELLOW}🔨 Building application...${NC}"
go build -o bin/it-bot-service .

# Crear datos de ejemplo si no existen
echo -e "${YELLOW}📊 Creating sample data...${NC}"
go run scripts/sample_data.go

# Iniciar el servicio
echo -e "${GREEN}🚀 Starting it-bot-service on port ${PORT:-8080}...${NC}"
echo -e "${GREEN}Health check: http://localhost:${PORT:-8080}/api/v1/health${NC}"
echo -e "${GREEN}API Documentation: http://localhost:${PORT:-8080}/swagger/index.html${NC}"
echo -e "${GREEN}Bot endpoints: http://localhost:${PORT:-8080}/api/v1/bots${NC}"
echo ""
echo -e "${BLUE}Press Ctrl+C to stop the service${NC}"
echo ""

# Ejecutar el servicio
./bin/it-bot-service