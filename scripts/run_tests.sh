#!/bin/bash

# run_tests.sh - Script para execução de testes do FHIR Guard CLI
# 
# Este script oferece várias opções para executar os diferentes 
# tipos de testes do projeto.

set -e

# Cores para melhor visualização
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Diretório raiz do projeto
ROOT_DIR=$(pwd)

print_header() {
    echo -e "\n${BLUE}=====================================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}=====================================================${NC}\n"
}

print_success() {
    echo -e "\n${GREEN}✓ $1${NC}\n"
}

print_error() {
    echo -e "\n${RED}✗ $1${NC}\n"
    exit 1
}

print_warning() {
    echo -e "\n${YELLOW}! $1${NC}\n"
}

run_unit_tests() {
    print_header "Executando testes unitários"
    go test -v -short ./cmd/... ./config/... || print_error "Testes unitários falharam"
    print_success "Testes unitários concluídos com sucesso"
}

run_integration_tests() {
    print_header "Executando testes de integração"
    go test -v ./... -run Integration -count=1 || print_error "Testes de integração falharam"
    print_success "Testes de integração concluídos com sucesso"
}

run_all_tests() {
    print_header "Executando todos os testes"
    go test -v ./... || print_error "Os testes falharam"
    print_success "Todos os testes concluídos com sucesso"
}

run_coverage() {
    print_header "Executando testes com cobertura"
    go test -v ./... -coverprofile=coverage.out || print_error "Testes de cobertura falharam"
    go tool cover -func=coverage.out
    
    # Gerar relatório HTML se solicitado
    if [ "$1" = "html" ]; then
        go tool cover -html=coverage.out -o coverage.html
        print_success "Relatório de cobertura HTML gerado em coverage.html"
        
        # Tenta abrir o relatório HTML automaticamente se disponível
        if command -v xdg-open >/dev/null 2>&1; then
            xdg-open coverage.html
        elif command -v open >/dev/null 2>&1; then
            open coverage.html
        else
            print_warning "Não foi possível abrir o relatório automaticamente. Abra manualmente o arquivo coverage.html"
        fi
    else
        print_success "Testes de cobertura concluídos com sucesso"
    fi
}

run_race_detector() {
    print_header "Executando testes com race detector"
    go test -v -race ./... || print_error "Race detector encontrou problemas"
    print_success "Testes com race detector concluídos com sucesso"
}

run_lint() {
    print_header "Executando verificação de lint"
    go vet ./... || print_error "go vet encontrou problemas"
    
    if command -v staticcheck >/dev/null 2>&1; then
        staticcheck ./... || print_error "staticcheck encontrou problemas"
    else
        print_warning "staticcheck não encontrado, pulando esta verificação"
        print_warning "Instale com: go install honnef.co/go/tools/cmd/staticcheck@latest"
    fi
    
    print_success "Verificação de lint concluída com sucesso"
}

run_check() {
    print_header "Executando verificação completa"
    run_lint
    run_unit_tests
    run_coverage
    print_success "Verificação completa concluída com sucesso"
}

show_help() {
    echo "Uso: ./scripts/run_tests.sh [opção]"
    echo ""
    echo "Opções:"
    echo "  unit          Executa apenas testes unitários"
    echo "  integration   Executa testes de integração"
    echo "  all           Executa todos os testes"
    echo "  coverage      Executa testes com relatório de cobertura"
    echo "  coverage-html Executa testes com relatório de cobertura HTML"
    echo "  race          Executa testes com race detector"
    echo "  lint          Executa verificação de lint"
    echo "  check         Executa verificação completa (lint + unit + coverage)"
    echo "  help          Mostra esta ajuda"
    echo ""
    echo "Sem argumentos, executa testes unitários."
}

# Verificar se o script está sendo executado no diretório raiz do projeto
if [ ! -f "go.mod" ]; then
    print_error "Este script deve ser executado no diretório raiz do projeto"
fi

# Criar diretório scripts se não existir
mkdir -p scripts

# Tornar o script executável
chmod +x scripts/run_tests.sh

# Processar argumentos
case "${1:-unit}" in
    unit)
        run_unit_tests
        ;;
    integration)
        run_integration_tests
        ;;
    all)
        run_all_tests
        ;;
    coverage)
        run_coverage
        ;;
    coverage-html)
        run_coverage html
        ;;
    race)
        run_race_detector
        ;;
    lint)
        run_lint
        ;;
    check)
        run_check
        ;;
    help)
        show_help
        ;;
    *)
        print_error "Opção desconhecida: $1. Use 'help' para ver as opções disponíveis."
        ;;
esac 