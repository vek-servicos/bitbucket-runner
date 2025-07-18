#!/bin/bash

# discord_send.sh - Script para envio de notificações ao Discord
# Versão: 2.0
# Melhorias implementadas:
# - Carregamento correto de dependências antes do trap
# - Função genérica de envio de mensagem
# - Tratamento de erros do Cloudflare
# - Informações do container no footer
# - Desativação de upload de arquivo e menções

set -euo pipefail

# Ativar debug conforme solicitado
set -x

# CORREÇÃO CRÍTICA: Carregar dependências ANTES do trap setup
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" &> /dev/null && pwd)"
source "${SCRIPT_DIR}/lib/common.sh"

# Agora configurar trap - as funções já estão disponíveis
trap 'handle_script_error $? $LINENO' ERR

# Configurações Discord
DISCORD_WEBHOOK_URL="${DISCORD_WEBHOOK_URL:-}"
CLOUDFLARE_API_TOKEN="${CLOUDFLARE_API_TOKEN:-}"
CLOUDFLARE_ZONE_ID="${CLOUDFLARE_ZONE_ID:-}"

# Função para tratamento de erros do script
handle_script_error() {
    local exit_code=$1
    local line_number=$2
    
    log "ERROR" "Script falhou na linha $line_number com código $exit_code"
    
    # Enviar notificação de erro crítico ao Discord
    send_error_notification "Falha crítica no script discord_send.sh" \
        "Linha: $line_number, Código: $exit_code"
}

# Função genérica para envio de mensagem ao Discord
# Parâmetros: título, descrição, cor (opcional), tipo (opcional)
send_discord_message() {
    local title="$1"
    local description="$2"
    local color="${3:-3447003}"  # Azul padrão
    local message_type="${4:-info}"
    
    if [[ -z "$DISCORD_WEBHOOK_URL" ]]; then
        log "WARN" "DISCORD_WEBHOOK_URL não configurado"
        return 1
    fi
    
    # Obter informações do ambiente
    local build_number=$(get_build_number)
    local timestamp=$(get_timestamp)
    local container_info=$(get_container_info)
    
    # Construir payload JSON
    local payload=$(cat <<EOF
{
    "embeds": [{
        "title": "$title",
        "description": "$description",
        "color": $color,
        "footer": {
            "text": "#$build_number • Bitbucket • $timestamp $container_info"
        },
        "timestamp": "$(date -u +%Y-%m-%dT%H:%M:%S.000Z)"
    }]
}
EOF
)
    
    # Enviar mensagem com retry
    local max_retries=3
    local retry_count=0
    
    while [[ $retry_count -lt $max_retries ]]; do
        if curl -s -X POST \
            -H "Content-Type: application/json" \
            -d "$payload" \
            "$DISCORD_WEBHOOK_URL" >/dev/null; then
            
            log "SUCCESS" "Mensagem enviada ao Discord: $title"
            return 0
        else
            ((retry_count++))
            log "WARN" "Tentativa $retry_count/$max_retries falhou"
            [[ $retry_count -lt $max_retries ]] && sleep 2
        fi
    done
    
    log "ERROR" "Falha ao enviar mensagem ao Discord após $max_retries tentativas"
    return 1
}

# Função específica para notificações de erro
send_error_notification() {
    local error_title="$1"
    local error_details="$2"
    
    log "INFO" "Enviando notificação de erro: $error_title"
    
    # Cor vermelha para erros
    send_discord_message "❌ $error_title" "$error_details" "15158332" "error"
}

# Função para verificar dependências
ensure_dependencies() {
    local missing_deps=()
    
    # Lista de dependências necessárias
    local required_commands=("curl" "jq")
    
    for cmd in "${required_commands[@]}"; do
        if ! command_exists "$cmd"; then
            missing_deps+=("$cmd")
        fi
    done
    
    if [[ ${#missing_deps[@]} -gt 0 ]]; then
        log "WARN" "Dependências ausentes: ${missing_deps[*]}"
        
        # Tentar instalar dependências baseado no sistema
        if command_exists "apt-get"; then
            log "INFO" "Instalando dependências via apt-get..."
            apt-get update -qq && apt-get install -y "${missing_deps[@]}"
        elif command_exists "yum"; then
            log "INFO" "Instalando dependências via yum..."
            yum install -y "${missing_deps[@]}"
        elif command_exists "apk"; then
            log "INFO" "Instalando dependências via apk..."
            apk add --no-cache "${missing_deps[@]}"
        else
            log "ERROR" "Não foi possível instalar dependências automaticamente"
            return 1
        fi
    fi
}

# Função para teste de conectividade com Cloudflare
test_cloudflare_connectivity() {
    if [[ -z "$CLOUDFLARE_API_TOKEN" ]] || [[ -z "$CLOUDFLARE_ZONE_ID" ]]; then
        log "WARN" "Credenciais do Cloudflare não configuradas"
        return 0
    fi
    
    log "INFO" "Testando conectividade com Cloudflare..."
    
    local response
    response=$(curl -s -w "%{http_code}" \
        -H "Authorization: Bearer $CLOUDFLARE_API_TOKEN" \
        -H "Content-Type: application/json" \
        "https://api.cloudflare.com/client/v4/zones/$CLOUDFLARE_ZONE_ID" \
        -o /dev/null)
    
    if [[ "$response" != "200" ]]; then
        local error_msg="Falha na comunicação com Cloudflare API (HTTP: $response)"
        log "ERROR" "$error_msg"
        
        # IMPLEMENTAÇÃO CRÍTICA: Notificar erro do Cloudflare via Discord
        send_error_notification "Cloudflare API Error" "$error_msg"
        return 1
    fi
    
    log "SUCCESS" "Conectividade com Cloudflare OK"
    return 0
}

# Função para envio de notificação de pipeline
send_pipeline_notification() {
    local pipeline_status="$1"
    local pipeline_title="${2:-Pipeline Notification}"
    
    case "$pipeline_status" in
        "success")
            send_discord_message "✅ $pipeline_title" "Pipeline executado com sucesso" "3066993"
            ;;
        "failure")
            send_discord_message "❌ $pipeline_title" "Pipeline falhou" "15158332"
            ;;
        "running")
            send_discord_message "🚀 $pipeline_title" "Pipeline em execução" "16776960"
            ;;
        *)
            send_discord_message "ℹ️ $pipeline_title" "Status: $pipeline_status" "3447003"
            ;;
    esac
}

# Função de ajuda
show_help() {
    cat <<EOF
Uso: $0 [OPÇÕES]

OPÇÕES:
    pipeline-notify STATUS TITLE    Enviar notificação de pipeline
    error-notify TITLE DETAILS      Enviar notificação de erro
    test-cloudflare                 Testar conectividade com Cloudflare
    --help                          Mostrar esta ajuda

EXEMPLOS:
    $0 pipeline-notify success "Deploy Production"
    $0 error-notify "Build Failed" "Compilation error in module X"
    $0 test-cloudflare

VARIÁVEIS DE AMBIENTE:
    DISCORD_WEBHOOK_URL     URL do webhook do Discord (obrigatório)
    CLOUDFLARE_API_TOKEN    Token da API do Cloudflare (opcional)
    CLOUDFLARE_ZONE_ID      ID da zona do Cloudflare (opcional)
EOF
}

# Função principal
main() {
    log "INFO" "Iniciando discord_send.sh v2.0"
    
    # Verificar dependências
    ensure_dependencies
    
    # Processar argumentos
    case "${1:-}" in
        "pipeline-notify")
            [[ $# -lt 3 ]] && { log "ERROR" "Argumentos insuficientes"; show_help; exit 1; }
            send_pipeline_notification "$2" "$3"
            ;;
        "error-notify")
            [[ $# -lt 3 ]] && { log "ERROR" "Argumentos insuficientes"; show_help; exit 1; }
            send_error_notification "$2" "$3"
            ;;
        "test-cloudflare")
            test_cloudflare_connectivity
            ;;
        "--help"|"-h"|"")
            show_help
            ;;
        *)
            log "ERROR" "Opção inválida: $1"
            show_help
            exit 1
            ;;
    esac
    
    log "SUCCESS" "discord_send.sh executado com sucesso"
}

# Executar função principal se script for chamado diretamente
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi