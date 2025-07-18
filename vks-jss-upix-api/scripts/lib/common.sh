#!/bin/bash

# common.sh - Biblioteca de funções comuns para scripts CI/CD
# Autor: Sistema de CI/CD 
# Descrição: Funções compartilhadas para logging, configuração e utilities

set -euo pipefail

# Cores para output
readonly RED='\033[0;31m'
readonly GREEN='\033[0;32m'
readonly YELLOW='\033[1;33m'
readonly BLUE='\033[0;34m'
readonly NC='\033[0m' # No Color

# Configurações globais
readonly SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" &> /dev/null && pwd)"
readonly PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../.." &> /dev/null && pwd)"

# Função de logging
log() {
    local level="$1"
    shift
    local message="$*"
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    
    case "$level" in
        "INFO")
            echo -e "${BLUE}[INFO]${NC} ${timestamp} - $message" >&2
            ;;
        "WARN")
            echo -e "${YELLOW}[WARN]${NC} ${timestamp} - $message" >&2
            ;;
        "ERROR")
            echo -e "${RED}[ERROR]${NC} ${timestamp} - $message" >&2
            ;;
        "SUCCESS")
            echo -e "${GREEN}[SUCCESS]${NC} ${timestamp} - $message" >&2
            ;;
        *)
            echo -e "[LOG] ${timestamp} - $message" >&2
            ;;
    esac
}

# Função para verificar se comando existe
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Função para verificar variáveis obrigatórias
require_env() {
    local var_name="$1"
    local var_value="${!var_name:-}"
    
    if [[ -z "$var_value" ]]; then
        log "ERROR" "Variável obrigatória $var_name não definida"
        return 1
    fi
}

# Função para cleanup em caso de erro
cleanup_on_error() {
    log "ERROR" "Script falhou. Executando cleanup..."
    # Adicionar lógica de cleanup aqui se necessário
}

# Função para obter informações do container
get_container_info() {
    if [[ -n "${BITBUCKET_DOCKER_IMAGE:-}" ]]; then
        echo "🐳 ${BITBUCKET_DOCKER_IMAGE}"
    elif [[ -f /.dockerenv ]]; then
        echo "🐳 $(hostname)"
    else
        echo "💻 $(whoami)@$(hostname):$(uname -s)"
    fi
}

# Função para obter build number
get_build_number() {
    echo "${BITBUCKET_BUILD_NUMBER:-$(date +%s)}"
}

# Função para obter timestamp formatado
get_timestamp() {
    date '+%Y-%m-%d %H:%M'
}

# Configurar trap para cleanup em caso de erro
trap cleanup_on_error ERR

log "INFO" "common.sh carregado com sucesso"