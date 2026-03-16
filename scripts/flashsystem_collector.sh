#!/bin/bash
# Script de wrapper para el colector de monitoreo IBM FlashSystem
# Este script actúa como intermediario entre Zabbix y el binario Go
# Proporciona seguridad adicional y manejo de errores para la integración con Zabbix

# Configuración del entorno
readonly SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
readonly COLLECTOR_BIN="$PROJECT_ROOT/flashsystem_collector"
readonly LOG_FILE="/var/log/zabbix/flashsystem_collector.log"
readonly CONFIG_DIR="/etc/zabbix/flashsystem"

# Función de logging para registrar eventos importantes
log_message() {
    local message="$1"
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    echo "[$timestamp] [FLASHSYSTEM_COLLECTOR] $message" >> "$LOG_FILE"
}

# Función para validar número de argumentos
validate_arguments() {
    local expected_count="$1"
    local actual_count="$#"
    
    if [ "$actual_count" -lt "$expected_count" ]; then
        log_message "ERROR: Número insuficiente de argumentos. Se esperaban al menos $expected_count, se recibieron $#"
        echo "0"
        exit 1
    fi
}

# Función para validar que el binario exista y tenga permisos de ejecución
validate_binary() {
    if [ ! -f "$COLLECTOR_BIN" ]; then
        log_message "ERROR: Binario no encontrado: $COLLECTOR_BIN"
        echo "0"
        exit 1
    fi
    
    if [ ! -x "$COLLECTOR_BIN" ]; then
        log_message "ERROR: Binario no tiene permisos de ejecución: $COLLECTOR_BIN"
        echo "0"
        exit 1
    fi
}

# Función para validar que el archivo de log sea escribible
validate_log_file() {
    # Crear directorio de logs si no existe
    local log_dir=$(dirname "$LOG_FILE")
    if [ ! -d "$log_dir" ]; thenp "$log_dir" 2>/dev/null || {
            # Si no se puede crear el directorio, usar un log temporal
            LOG_FILE="/tmp/flashsystem_collector.log"
        }
    fi
    
    # Verificar que el archivo de log sea escribible
    if [ ! -w "$log_dir" ]; then
        log_message "WARNING: Directorio de logs no escribible, usando /tmp"
        LOG_FILE="/tmp/flashsystem_collector.log"
    fi
}

# Función para ejecutar el colector con timeout y manejo de errores
execute_collector() {
    local timeout_duration=60  # 60 segundos de timeout por defecto
    
    # Usar timeout para prevenir que el proceso se quede colgado
    if command -v timeout >/dev/null 2>&1; then
        # Ejecutar con timeout si el comando está disponible
        timeout "$timeout_duration" "$COLLECTOR_BIN" "$@" 2>>"$LOG_FILE"
        local exit_code=$?
        
        # Verificar si el timeout mató al proceso
        if [ $exit_code -eq 124 ]; then
            log_message "ERROR: Timeout ejecutando colector con argumentos: $1 $2 $3 $4"
            echo "0"
            exit 1
        fi
        
        return $exit_code
    else
        # Si timeout no está disponible, ejecutar sin él
        "$COLLECTOR_BIN" "$@" 2>>"$LOG_FILE"
        return $?
    fi
}

# Función principal de inicialización
initialize() {
    # Validar que todos los parámetros necesarios estén presentes
    validate_arguments 4 "$@"
    
    # Validar que el binario exista y sea ejecutable
    validate_binary
    
    # Validar que el archivo de log sea escribible
    validate_log_file
    
    # Registrar inicio de ejecución
    log_message "Iniciando ejecución con argumentos: $1 *** *** $4 [$(echo "$@" | cut -d' ' -f5-)]"
}

# Manejar señales para limpieza adecuada
trap cleanup_handler SIGTERM SIGINT

cleanup_handler() {
    log_message "Recepción de señal, terminando ejecución"
    exit 0
}

# Verificar que se hayan pasado argumentos
if [ $# -lt 4 ]; then
    echo "0"
    exit 1
fi

# Inicializar el script
initialize "$@"

# Ejecutar el colector con los argumentos pasados
execute_collector "$@"

# Capturar el código de salida
exit_code=$?

# Registrar resultado de la ejecución
if [ $exit_code -ne 0 ]; then
    log_message "ERROR: Colector terminó con código $exit_code"
else
    log_message "SUCCESS: Colector terminó exitosamente"
fi

# Salir con el mismo código de salida del colector
exit $exit_code