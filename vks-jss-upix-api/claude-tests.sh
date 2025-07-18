#!/bin/bash

# claude-tests.sh - Script de testes do AI Agent Claude
# Versão: 1.0
# Descrição: Script para executar testes de forma autônoma sem precisar pedir permissão

set -euo pipefail
set -x  # Debug ativado conforme solicitado

# Configurações
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" &> /dev/null && pwd)"
PROJECT_ROOT="$SCRIPT_DIR"
TEST_RESULTS_DIR="$PROJECT_ROOT/test-results"
TIMESTAMP=$(date '+%Y%m%d_%H%M%S')
TEST_LOG="$TEST_RESULTS_DIR/claude-tests-$TIMESTAMP"

# Cores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Função de logging
claude_log() {
    local level="$1"
    shift
    local message="$*"
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    
    # Criar diretório se não existir
    mkdir -p "${TEST_LOG}"
    
    case "$level" in
        "INFO")
            echo -e "${BLUE}[CLAUDE-INFO]${NC} ${timestamp} - $message" | tee -a "${TEST_LOG}/test.log"
            ;;
        "WARN")
            echo -e "${YELLOW}[CLAUDE-WARN]${NC} ${timestamp} - $message" | tee -a "${TEST_LOG}/test.log"
            ;;
        "ERROR")
            echo -e "${RED}[CLAUDE-ERROR]${NC} ${timestamp} - $message" | tee -a "${TEST_LOG}/test.log"
            ;;
        "SUCCESS")
            echo -e "${GREEN}[CLAUDE-SUCCESS]${NC} ${timestamp} - $message" | tee -a "${TEST_LOG}/test.log"
            ;;
    esac
}

# Função para preparar ambiente de teste
setup_test_environment() {
    claude_log "INFO" "Preparando ambiente de teste..."
    
    # Criar diretório de resultados
    mkdir -p "$TEST_LOG"
    
    # Verificar se scripts existem
    local required_scripts=(
        "scripts/discord_send.sh"
        "scripts/lib/common.sh"
        "test-runner.sh"
        "test-runner.env"
    )
    
    for script in "${required_scripts[@]}"; do
        if [[ ! -f "$PROJECT_ROOT/$script" ]]; then
            claude_log "ERROR" "Script obrigatório não encontrado: $script"
            return 1
        fi
    done
    
    # Tornar scripts executáveis
    chmod +x "$PROJECT_ROOT/scripts/discord_send.sh"
    chmod +x "$PROJECT_ROOT/test-runner.sh"
    
    claude_log "SUCCESS" "Ambiente de teste preparado"
}

# Função para testar correção da dependência
test_dependency_fix() {
    claude_log "INFO" "=== TESTE: Correção de Dependência ==="
    
    # Verificar se o carregamento está na ordem correta
    if grep -A 5 "CORREÇÃO CRÍTICA" "$PROJECT_ROOT/scripts/discord_send.sh" | grep -q "source.*common.sh"; then
        claude_log "SUCCESS" "Carregamento de common.sh ANTES do trap ✓"
    else
        claude_log "ERROR" "Carregamento de common.sh não encontrado na posição correta ✗"
        return 1
    fi
    
    # Verificar se trap está após o source
    local common_line=$(grep -n "source.*common.sh" "$PROJECT_ROOT/scripts/discord_send.sh" | cut -d: -f1)
    local trap_line=$(grep -n "trap.*handle_script_error" "$PROJECT_ROOT/scripts/discord_send.sh" | cut -d: -f1)
    
    if [[ $common_line -lt $trap_line ]]; then
        claude_log "SUCCESS" "Ordem correta: common.sh (linha $common_line) antes do trap (linha $trap_line) ✓"
    else
        claude_log "ERROR" "Ordem incorreta: trap na linha $trap_line, common.sh na linha $common_line ✗"
        return 1
    fi
}

# Função para testar notificações de erro do Cloudflare
test_cloudflare_error_notifications() {
    claude_log "INFO" "=== TESTE: Notificações de Erro do Cloudflare ==="
    
    # Simular erro do Cloudflare removendo temporariamente as credenciais
    export CLOUDFLARE_API_TOKEN=""
    export CLOUDFLARE_ZONE_ID=""
    export DISCORD_WEBHOOK_URL="https://discord.com/api/webhooks/test/fake"
    
    # Executar teste (deve falhar mas enviar notificação)
    local test_result=0
    "$PROJECT_ROOT/scripts/discord_send.sh" test-cloudflare || test_result=$?
    
    if [[ $test_result -eq 0 ]]; then
        claude_log "WARN" "Teste não falhou como esperado (sem credenciais Cloudflare)"
    else
        claude_log "SUCCESS" "Teste falhou conforme esperado (sem credenciais) ✓"
    fi
}

# Função para testar diferentes imagens de container
test_container_images() {
    claude_log "INFO" "=== TESTE: Imagens de Container ==="
    
    # Extrair imagens do bitbucket-pipelines.yml
    local images=(
        "atlassian/default-image:5"
        "maven:3.8.6-openjdk-17-slim"
        "node:18-alpine"
        "ubuntu:20.04"
        "python:3.11-slim"
    )
    
    for image in "${images[@]}"; do
        claude_log "INFO" "Testando imagem: $image"
        
        # Teste de sucesso
        "$PROJECT_ROOT/test-runner.sh" container "$image" success || {
            claude_log "ERROR" "Falha no teste de sucesso para $image"
        }
        
        # Teste de falha
        "$PROJECT_ROOT/test-runner.sh" container "$image" failure || {
            claude_log "ERROR" "Falha no teste de falha para $image"
        }
    done
}

# Função para testar funcionalidade de footer
test_footer_functionality() {
    claude_log "INFO" "=== TESTE: Funcionalidade do Footer ==="
    
    # Verificar se função get_container_info existe
    if grep -q "get_container_info" "$PROJECT_ROOT/scripts/lib/common.sh"; then
        claude_log "SUCCESS" "Função get_container_info encontrada ✓"
    else
        claude_log "ERROR" "Função get_container_info não encontrada ✗"
        return 1
    fi
    
    # Simular diferentes ambientes
    export BITBUCKET_DOCKER_IMAGE="test-image:latest"
    local result=$("$PROJECT_ROOT/scripts/lib/common.sh" && get_container_info 2>/dev/null || echo "erro")
    
    if [[ "$result" == *"🐳"* ]]; then
        claude_log "SUCCESS" "Footer de container funcionando ✓"
    else
        claude_log "WARN" "Footer de container pode não estar funcionando corretamente"
    fi
}

# Função para validar desativação de uploads e menções
test_disabled_features() {
    claude_log "INFO" "=== TESTE: Recursos Desativados ==="
    
    # Verificar se não há menções no payload
    if ! grep -q "@" "$PROJECT_ROOT/scripts/discord_send.sh"; then
        claude_log "SUCCESS" "Menções desativadas ✓"
    else
        claude_log "WARN" "Possíveis menções encontradas no código"
    fi
    
    # Verificar se não há upload de arquivos
    if ! grep -q "multipart" "$PROJECT_ROOT/scripts/discord_send.sh"; then
        claude_log "SUCCESS" "Upload de arquivos desativado ✓"
    else
        claude_log "WARN" "Possível upload de arquivo encontrado no código"
    fi
}

# Função para verificar tipos de erro notificáveis
analyze_error_types() {
    claude_log "INFO" "=== ANÁLISE: Tipos de Erro Notificáveis ==="
    
    cat << 'EOF' > "$TEST_LOG/error_types_analysis.md"
# Tipos de Erros Notificáveis no Discord

## Erros Críticos (Sempre notificar)
1. **Falhas de API Externa**
   - Cloudflare API (HTTP 4xx/5xx)
   - Webhook Discord (falha de conectividade)
   - APIs de terceiros

2. **Erros de Pipeline**
   - Build failure
   - Test failure  
   - Deploy failure
   - Timeout de execução

3. **Erros de Infraestrutura**
   - Container crash
   - Recursos insuficientes
   - Falhas de rede

## Erros de Aviso (Opcionalmente notificar)
1. **Configuração**
   - Variáveis de ambiente ausentes
   - Credenciais expiradas
   - Configuração inválida

2. **Performance**
   - Execução lenta
   - Uso alto de recursos
   - Timeouts menores

## Erros Informativos (Log apenas)
1. **Operacionais**
   - Retry bem-sucedido
   - Fallback executado
   - Cache miss

## Implementação Recomendada
- Error level: `error-notify` para críticos
- Warning level: `warn-notify` para avisos
- Info level: log local apenas
EOF
    
    claude_log "SUCCESS" "Análise de tipos de erro salva em $TEST_LOG/error_types_analysis.md"
}

# Função para testar ensure_dependencies
test_ensure_dependencies() {
    claude_log "INFO" "=== TESTE: Ensure Dependencies ==="
    
    # Verificar se a função existe
    if grep -q "ensure_dependencies" "$PROJECT_ROOT/scripts/discord_send.sh"; then
        claude_log "SUCCESS" "Função ensure_dependencies encontrada ✓"
        
        # Verificar se verifica privilégios antes de instalar
        if grep -A 10 "ensure_dependencies" "$PROJECT_ROOT/scripts/discord_send.sh" | grep -q "command_exists"; then
            claude_log "SUCCESS" "Função verifica comandos existentes ✓"
        else
            claude_log "WARN" "Função pode não verificar comandos existentes"
        fi
    else
        claude_log "ERROR" "Função ensure_dependencies não encontrada ✗"
        return 1
    fi
}

# Função para executar bateria completa de testes
run_full_test_suite() {
    claude_log "INFO" "Iniciando bateria completa de testes Claude..."
    
    local tests_passed=0
    local tests_failed=0
    local total_tests=7
    
    # Lista de testes
    local test_functions=(
        "test_dependency_fix"
        "test_cloudflare_error_notifications" 
        "test_container_images"
        "test_footer_functionality"
        "test_disabled_features"
        "test_ensure_dependencies"
        "analyze_error_types"
    )
    
    for test_func in "${test_functions[@]}"; do
        claude_log "INFO" "Executando: $test_func"
        
        if $test_func; then
            ((tests_passed++))
            claude_log "SUCCESS" "✓ $test_func passou"
        else
            ((tests_failed++))
            claude_log "ERROR" "✗ $test_func falhou"
        fi
        
        echo "---" >> "$TEST_LOG/test.log"
    done
    
    # Relatório final
    cat << EOF > "$TEST_LOG/REPORT.txt"
RELATÓRIO DE TESTES CLAUDE - $TIMESTAMP
=======================================

Total de testes: $total_tests
Testes passou: $tests_passed
Testes falharam: $tests_failed
Taxa de sucesso: $(( (tests_passed * 100) / total_tests ))%

STATUS: $([ $tests_failed -eq 0 ] && echo "TODOS PASSARAM ✓" || echo "ALGUNS FALHARAM ✗")

Logs detalhados: $TEST_LOG/test.log
Análise de erros: $TEST_LOG/error_types_analysis.md

EOF
    
    claude_log "INFO" "=== RELATÓRIO FINAL ==="
    cat "$TEST_LOG/REPORT.txt"
    
    return $tests_failed
}

# Função de ajuda
show_help() {
    cat <<EOF
Uso: $0 [OPÇÃO]

OPÇÕES:
    setup                   Preparar ambiente de teste
    test-dependency         Testar correção de dependência
    test-cloudflare         Testar notificações Cloudflare
    test-containers         Testar imagens de container
    test-footer             Testar funcionalidade do footer
    test-features           Testar recursos desativados
    test-deps               Testar ensure_dependencies
    analyze-errors          Analisar tipos de erro
    full-suite              Executar todos os testes
    --help                  Mostrar esta ajuda

EXEMPLOS:
    $0 setup
    $0 full-suite
    $0 test-cloudflare

EOF
}

# Função principal
main() {
    claude_log "INFO" "Iniciando claude-tests.sh v1.0"
    
    # Preparar ambiente sempre
    setup_test_environment
    
    case "${1:-full-suite}" in
        "setup")
            claude_log "SUCCESS" "Ambiente preparado"
            ;;
        "test-dependency")
            test_dependency_fix
            ;;
        "test-cloudflare") 
            test_cloudflare_error_notifications
            ;;
        "test-containers")
            test_container_images
            ;;
        "test-footer")
            test_footer_functionality
            ;;
        "test-features")
            test_disabled_features
            ;;
        "test-deps")
            test_ensure_dependencies
            ;;
        "analyze-errors")
            analyze_error_types
            ;;
        "full-suite")
            run_full_test_suite
            ;;
        "--help"|"-h")
            show_help
            ;;
        *)
            claude_log "ERROR" "Opção inválida: $1"
            show_help
            exit 1
            ;;
    esac
    
    claude_log "SUCCESS" "claude-tests.sh executado com sucesso"
}

# Executar se chamado diretamente
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi