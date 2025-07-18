#!/bin/bash

# test-runner.sh - Script para simulação de ambiente Bitbucket Pipelines
# Versão: 1.0
# Descrição: Simula ambiente do Bitbucket para testar scripts CI/CD

set -euo pipefail

# Ativar debug
set -x

# Configurações
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" &> /dev/null && pwd)"
PROJECT_ROOT="$SCRIPT_DIR"

# Cores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Função de logging
test_log() {
    local level="$1"
    shift
    local message="$*"
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    
    case "$level" in
        "INFO")
            echo -e "${BLUE}[TEST-INFO]${NC} ${timestamp} - $message"
            ;;
        "WARN")
            echo -e "${YELLOW}[TEST-WARN]${NC} ${timestamp} - $message"
            ;;
        "ERROR")
            echo -e "${RED}[TEST-ERROR]${NC} ${timestamp} - $message"
            ;;
        "SUCCESS")
            echo -e "${GREEN}[TEST-SUCCESS]${NC} ${timestamp} - $message"
            ;;
    esac
}

# Função para carregar configurações
load_test_env() {
    local env_file="$PROJECT_ROOT/test-runner.env"
    
    if [[ -f "$env_file" ]]; then
        test_log "INFO" "Carregando variáveis de $env_file"
        set -a  # Exportar automaticamente as variáveis
        source "$env_file"
        set +a
    else
        test_log "ERROR" "Arquivo $env_file não encontrado"
        return 1
    fi
}

# Função para carregar mapeamento de imagens
load_image_mapping() {
    local mapping_file="$PROJECT_ROOT/scripts/config/image-mapping.conf"
    
    if [[ -f "$mapping_file" ]]; then
        test_log "INFO" "Carregando mapeamento de imagens de $mapping_file"
        source "$mapping_file"
    else
        test_log "WARN" "Arquivo de mapeamento $mapping_file não encontrado"
    fi
}

# Função para simular ambiente Docker
simulate_docker_env() {
    local image_name="$1"
    
    test_log "INFO" "Simulando ambiente Docker: $image_name"
    
    # Definir variável da imagem
    export BITBUCKET_DOCKER_IMAGE="$image_name"
    
    # Simular arquivo .dockerenv se não existir
    if [[ ! -f /.dockerenv ]]; then
        test_log "INFO" "Criando simulação de ambiente Docker"
        touch /tmp/.dockerenv.sim
        export DOCKER_ENV_SIM="/tmp/.dockerenv.sim"
    fi
}

# Função para executar teste em container específico
run_container_test() {
    local container_image="$1"
    local test_scenario="$2"
    local expected_result="$3"
    
    test_log "INFO" "=== TESTE: $test_scenario - Imagem: $container_image ==="
    
    # Simular ambiente do container
    simulate_docker_env "$container_image"
    
    # Definir cenário de teste
    case "$test_scenario" in
        "success")
            export PIPELINE_STATUS="success"
            export CLOUDFLARE_API_TOKEN="valid_token"
            ;;
        "failure")
            export PIPELINE_STATUS="failure"
            export CLOUDFLARE_API_TOKEN="invalid_token"
            ;;
        "cloudflare_error")
            export CLOUDFLARE_API_TOKEN=""
            ;;
        *)
            test_log "WARN" "Cenário desconhecido: $test_scenario"
            ;;
    esac
    
    # Executar teste
    local test_cmd="$PROJECT_ROOT/scripts/discord_send.sh"
    local test_result=0
    
    case "$test_scenario" in
        "success"|"failure")
            $test_cmd pipeline-notify "$PIPELINE_STATUS" "Test Pipeline - $container_image" || test_result=$?
            ;;
        "cloudflare_error")
            $test_cmd test-cloudflare || test_result=$?
            ;;
    esac
    
    # Verificar resultado
    if [[ "$expected_result" == "should_pass" && $test_result -eq 0 ]]; then
        test_log "SUCCESS" "Teste passou conforme esperado"
        return 0
    elif [[ "$expected_result" == "should_fail" && $test_result -ne 0 ]]; then
        test_log "SUCCESS" "Teste falhou conforme esperado"
        return 0
    else
        test_log "ERROR" "Resultado inesperado. Esperado: $expected_result, Código: $test_result"
        return 1
    fi
}

# Função para executar bateria completa de testes
run_full_test_suite() {
    test_log "INFO" "Iniciando bateria completa de testes"
    
    local total_tests=0
    local passed_tests=0
    local failed_tests=0
    
    # Lista de imagens para testar (do image-mapping.conf)
    local images=("$default" "$ubuntu" "$node" "$python")
    
    for image in "${images[@]}"; do
        test_log "INFO" "Testando imagem: $image"
        
        # Teste de sucesso
        ((total_tests++))
        if run_container_test "$image" "success" "should_pass"; then
            ((passed_tests++))
        else
            ((failed_tests++))
        fi
        
        # Teste de falha
        ((total_tests++))
        if run_container_test "$image" "failure" "should_pass"; then
            ((passed_tests++))
        else
            ((failed_tests++))
        fi
        
        # Teste de erro do Cloudflare
        ((total_tests++))
        if run_container_test "$image" "cloudflare_error" "should_pass"; then
            ((passed_tests++))
        else
            ((failed_tests++))
        fi
    done
    
    # Relatório final
    test_log "INFO" "=== RELATÓRIO FINAL ==="
    test_log "INFO" "Total de testes: $total_tests"
    test_log "SUCCESS" "Testes passaram: $passed_tests"
    test_log "ERROR" "Testes falharam: $failed_tests"
    
    if [[ $failed_tests -eq 0 ]]; then
        test_log "SUCCESS" "Todos os testes passaram!"
        return 0
    else
        test_log "ERROR" "Alguns testes falharam"
        return 1
    fi
}

# Função de ajuda
show_help() {
    cat <<EOF
Uso: $0 [OPÇÕES]

OPÇÕES:
    container IMAGE TEST_TYPE       Executar teste específico em container
    full-suite                      Executar bateria completa de testes
    list-images                     Listar imagens disponíveis
    --help                          Mostrar esta ajuda

TIPOS DE TESTE:
    success                         Teste de pipeline com sucesso
    failure                         Teste de pipeline com falha
    cloudflare_error               Teste de erro do Cloudflare

EXEMPLOS:
    $0 container "atlassian/default-image:5" success
    $0 full-suite
    $0 list-images

EOF
}

# Função para listar imagens disponíveis
list_images() {
    test_log "INFO" "Imagens disponíveis para teste:"
    load_image_mapping
    
    echo "default: $default"
    echo "ubuntu: $ubuntu"
    echo "node: $node"
    echo "python: $python"
    echo "openjdk: $openjdk"
    echo "maven: $maven"
}

# Função principal
main() {
    test_log "INFO" "Iniciando test-runner.sh v1.0"
    
    # Carregar configurações
    load_test_env
    load_image_mapping
    
    # Processar argumentos
    case "${1:-}" in
        "container")
            [[ $# -lt 3 ]] && { test_log "ERROR" "Argumentos insuficientes"; show_help; exit 1; }
            run_container_test "$2" "$3" "should_pass"
            ;;
        "full-suite")
            run_full_test_suite
            ;;
        "list-images")
            list_images
            ;;
        "--help"|"-h"|"")
            show_help
            ;;
        *)
            test_log "ERROR" "Opção inválida: $1"
            show_help
            exit 1
            ;;
    esac
}

# Cleanup no final
cleanup() {
    if [[ -f "/tmp/.dockerenv.sim" ]]; then
        rm -f "/tmp/.dockerenv.sim"
    fi
}

trap cleanup EXIT

# Executar função principal se script for chamado diretamente
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi