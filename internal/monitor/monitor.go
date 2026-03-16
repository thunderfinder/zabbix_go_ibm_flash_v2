// Package monitor implementa funciones específicas de monitoreo para componentes IBM FlashSystem
// Este paquete interactúa con el sistema de almacenamiento para obtener métricas específicas
package monitor

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"zabbix_go_ibm_flash_v2/internal/parser"
	"zabbix_go_ibm_flash_v2/internal/ssh"
)

// GetDriveStatus obtiene el estado de una unidad específica por enclosure y slot
// host: dirección IP o nombre de host del sistema FlashSystem
// user: nombre de usuario para autenticación
// pass: contraseña para autenticación
// enclosureID: ID del enclosure donde está la unidad
// slotID: ID del slot donde está la unidad
// Retorna 1 para online, 0 para offline, 2 para otros estados
func GetDriveStatus(host, user, pass, enclosureID, slotID string) (string, error) {
	// Validar entradas para prevenir inyección de comandos
	if !isValidNumericInput(enclosureID) || !isValidNumericInput(slotID) {
		return "0", fmt.Errorf("argumentos inválidos para GetDriveStatus: enclosure=%s, slot=%s", enclosureID, slotID)
	}

	// Establecer conexión SSH
	client, err := ssh.EstablishSSHConnection(host, user, pass)
	if err != nil {
		return "0", fmt.Errorf("error estableciendo conexión SSH: %w", err)
	}
	defer func() {
		if closeErr := client.Close(); closeErr != nil {
			log.Printf("Advertencia: Error al cerrar conexión SSH: %v", closeErr)
		}
	}()

	// Ejecutar comando para obtener información de unidades
	cmd := "svcinfo lsdrive -delim :"
	output, err := ssh.ExecuteCommand(client, cmd)
	if err != nil {
		return "0", fmt.Errorf("error ejecutando comando lsdrive: %w", err)
	}

	// Parsear la salida del comando
	records := parser.ParseDelimitedOutput(output, ':')

	// Buscar la unidad específica en los registros
	for _, r := range records {
		if r["enclosure_id"] == enclosureID && r["slot_id"] == slotID {
			// Convertir el estado a un valor numérico para Zabbix
			status := r["status"]
			switch strings.ToLower(status) {
			case "online", "available", "good", "active", "ready", "normal":
				return "1", nil // Estado OK
			case "offline", "unavailable", "failed", "inactive", "missing", "degraded", "error":
				return "0", nil // Estado ERROR
			default:
				log.Printf("Estado desconocido de unidad: %s", status)
				return "2", nil // Estado DESCONOCIDO
			}
		}
	}

	// Si no se encuentra la unidad, retornar error
	log.Printf("Unidad no encontrada: enclosure=%s, slot=%s", enclosureID, slotID)
	return "0", nil
}

// GetEnclosureStatus obtiene el estado de una carcasa específica por ID
// host: dirección IP o nombre de host del sistema FlashSystem
// user: nombre de usuario para autenticación
// pass: contraseña para autenticación
// enclosureID: ID de la carcasa a consultar
// Retorna 1 para online, 0 para offline
func GetEnclosureStatus(host, user, pass, enclosureID string) (string, error) {
	// Validar entrada para prevenir inyección de comandos
	if !isValidNumericInput(enclosureID) {
		return "0", fmt.Errorf("argumento inválido para GetEnclosureStatus: enclosure=%s", enclosureID)
	}

	// Establecer conexión SSH
	client, err := ssh.EstablishSSHConnection(host, user, pass)
	if err != nil {
		return "0", fmt.Errorf("error estableciendo conexión SSH: %w", err)
	}
	defer func() {
		if closeErr := client.Close(); closeErr != nil {
			log.Printf("Advertencia: Error al cerrar conexión SSH: %v", closeErr)
		}
	}()

	// Ejecutar comando para obtener información de carcasas
	cmd := "svcinfo lsenclosure -delim :"
	output, err := ssh.ExecuteCommand(client, cmd)
	if err != nil {
		return "0", fmt.Errorf("error ejecutando comando lsenclosure: %w", err)
	}

	// Parsear la salida del comando
	records := parser.ParseDelimitedOutput(output, ':')

	// Buscar la carcasa específica en los registros
	for _, r := range records {
		if r["id"] == enclosureID {
			// Convertir el estado a un valor numérico para Zabbix
			status := r["status"]
			switch strings.ToLower(status) {
			case "online", "active", "operational", "running", "normal":
				return "1", nil // Estado OK
			case "offline", "inactive", "failed", "stopped", "error", "degraded":
				return "0", nil // Estado ERROR
			default:
				log.Printf("Estado desconocido de carcasa: %s", status)
				return "2", nil // Estado DESCONOCIDO
			}
		}
	}

	// Si no se encuentra la carcasa, retornar error
	log.Printf("Carcasa no encontrada: id=%s", enclosureID)
	return "0", nil
}

// GetBatteryStatus obtiene el estado de la batería de una carcasa específica
// host: dirección IP o nombre de host del sistema FlashSystem
// user: nombre de usuario para autenticación
// pass: contraseña para autenticación
// enclosureID: ID de la carcasa a consultar
// Retorna 1 para online, 0 para offline
func GetBatteryStatus(host, user, pass, enclosureID string) (string, error) {
	// Validar entrada para prevenir inyección de comandos
	if !isValidNumericInput(enclosureID) {
		return "0", fmt.Errorf("argumento inválido para GetBatteryStatus: enclosure=%s", enclosureID)
	}

	// Establecer conexión SSH
	client, err := ssh.EstablishSSHConnection(host, user, pass)
	if err != nil {
		return "0", fmt.Errorf("error estableciendo conexión SSH: %w", err)
	}
	defer func() {
		if closeErr := client.Close(); closeErr != nil {
			log.Printf("Advertencia: Error al cerrar conexión SSH: %v", closeErr)
		}
	}()

	// Ejecutar comando para obtener información de baterías
	cmd := "svcinfo lsenclosurebattery -delim :"
	output, err := ssh.ExecuteCommand(client, cmd)
	if err != nil {
		return "0", fmt.Errorf("error ejecutando comando lsenclosurebattery: %w", err)
	}

	// Parsear la salida del comando
	records := parser.ParseDelimitedOutput(output, ':')

	// Buscar la batería de la carcasa específica en los registros
	for _, r := range records {
		if r["enclosure_id"] == enclosureID {
			// Convertir el estado a un valor numérico para Zabbix
			status := r["status"]
			switch strings.ToLower(status) {
			case "online", "present", "optimal", "good", "charging", "normal":
				return "1", nil // Estado OK
			case "offline", "absent", "failed", "critical", "low", "discharging", "error":
				return "0", nil // Estado ERROR
			default:
				log.Printf("Estado desconocido de batería: %s", status)
				return "2", nil // Estado DESCONOCIDO
			}
		}
	}

	// Si no se encuentra la batería, retornar error
	log.Printf("Batería no encontrada para carcasa: id=%s", enclosureID)
	return "0", nil
}

// GetPoolUsage obtiene el porcentaje de uso de un grupo de discos por nombre
// host: dirección IP o nombre de host del sistema FlashSystem
// user: nombre de usuario para autenticación
// pass: contraseña para autenticación
// poolName: nombre del grupo de discos a consultar
// Retorna un valor decimal representando el porcentaje de uso
func GetPoolUsage(host, user, pass, poolName string) (float64, error) {
	// Validar entrada para prevenir inyección de comandos
	poolName = strings.TrimSpace(poolName)
	if poolName == "" {
		return 0, fmt.Errorf("nombre de pool vacío proporcionado")
	}

	// Establecer conexión SSH
	client, err := ssh.EstablishSSHConnection(host, user, pass)
	if err != nil {
		return 0, fmt.Errorf("error estableciendo conexión SSH: %w", err)
	}
	defer func() {
		if closeErr := client.Close(); closeErr != nil {
			log.Printf("Advertencia: Error al cerrar conexión SSH: %v", closeErr)
		}
	}()

	// Ejecutar comando para obtener información de grupos de discos
	cmd := "svcinfo lsmdiskgrp -delim :"
	output, err := ssh.ExecuteCommand(client, cmd)
	if err != nil {
		return 0, fmt.Errorf("error ejecutando comando lsmdiskgrp: %w", err)
	}

	// Parsear la salida del comando
	records := parser.ParseDelimitedOutput(output, ':')

	// Buscar el grupo de discos específico en los registros
	for _, r := range records {
		if r["name"] == poolName {
			// Extraer valores numéricos del grupo de discos
			totalStr := strings.TrimSpace(r["capacity"])
			freeStr := strings.TrimSpace(r["free_capacity"])

			// Convertir cadenas a números decimales
			total, err := convertToFloat(totalStr)
			if err != nil {
				return 0, fmt.Errorf("valor de capacidad total inválido '%s': %w", totalStr, err)
			}

			free, err := convertToFloat(freeStr)
			if err != nil {
				return 0, fmt.Errorf("valor de capacidad libre inválido '%s': %w", freeStr, err)
			}

			// Validar que los valores sean razonables
			if total <= 0 {
				return 0, fmt.Errorf("capacidad total inválida: %f", total)
			}

			if free > total {
				return 0, fmt.Errorf("capacidad libre %f mayor que capacidad total %f", free, total)
			}

			// Calcular porcentaje de uso
			used := total - free
			percentage := (used / total) * 100

			return percentage, nil
		}
	}

	// Si no se encuentra el grupo de discos, retornar 0
	log.Printf("Grupo de discos no encontrado: %s", poolName)
	return 0, nil
}

// GetTotalIOPS obtiene las operaciones de entrada/salida por segundo del sistema
// host: dirección IP o nombre de host del sistema FlashSystem
// user: nombre de usuario para autenticación
// pass: contraseña para autenticación
// Retorna un valor entero representando las IOPS totales
func GetTotalIOPS(host, user, pass string) (int, error) {
	// Establecer conexión SSH
	client, err := ssh.EstablishSSHConnection(host, user, pass)
	if err != nil {
		return 0, fmt.Errorf("error estableciendo conexión SSH: %w", err)
	}
	defer func() {
		if closeErr := client.Close(); closeErr != nil {
			log.Printf("Advertencia: Error al cerrar conexión SSH: %v", closeErr)
		}
	}()

	// Ejecutar comando para obtener estadísticas del sistema
	cmd := "svcinfo lssystemstats -delim :"
	output, err := ssh.ExecuteCommand(client, cmd)
	if err != nil {
		return 0, fmt.Errorf("error ejecutando comando lssystemstats: %w", err)
	}

	// Parsear la salida del comando
	records := parser.ParseDelimitedOutput(output, ':')

	// El comando lssystemstats normalmente devuelve una sola fila
	for _, r := range records {
		// Extraer valores de IOPS de lectura y escritura
		readIOPS, err1 := convertToInt(r["read_io"])
		writeIOPS, err2 := convertToInt(r["write_io"])

		// Si ambos valores son válidos, sumarlos
		if err1 == nil && err2 == nil {
			return readIOPS + writeIOPS, nil
		}

		// Si solo uno es válido, usar ese
		if err1 == nil {
			return readIOPS, nil
		}
		if err2 == nil {
			return writeIOPS, nil
		}
	}

	// Si no se puede obtener IOPS, retornar 0
	log.Println("No se pudo obtener IOPS del sistema")
	return 0, nil
}

// GetVolumeStatus obtiene el estado de un volumen específico por nombre
// host: dirección IP o nombre de host del sistema FlashSystem
// user: nombre de usuario para autenticación
// pass: contraseña para autenticación
// volumeName: nombre del volumen a consultar
// Retorna 1 para online, 0 para offline, 2 para otros estados
func GetVolumeStatus(host, user, pass, volumeName string) (string, error) {
	// Validar entrada para prevenir inyección de comandos
	volumeName = strings.TrimSpace(volumeName)
	if volumeName == "" {
		return "0", fmt.Errorf("nombre de volumen vacío proporcionado")
	}

	// Establecer conexión SSH
	client, err := ssh.EstablishSSHConnection(host, user, pass)
	if err != nil {
		return "0", fmt.Errorf("error estableciendo conexión SSH: %w", err)
	}
	defer func() {
		if closeErr := client.Close(); closeErr != nil {
			log.Printf("Advertencia: Error al cerrar conexión SSH: %v", closeErr)
		}
	}()

	// Ejecutar comando para obtener información de volúmenes
	cmd := "svcinfo lsvdisk -delim :"
	output, err := ssh.ExecuteCommand(client, cmd)
	if err != nil {
		return "0", fmt.Errorf("error ejecutando comando lsvdisk: %w", err)
	}

	// Parsear la salida del comando
	records := parser.ParseDelimitedOutput(output, ':')

	// Buscar el volumen específico en los registros
	for _, r := range records {
		if r["name"] == volumeName {
			// Convertir el estado a un valor numérico para Zabbix
			status := r["status"]
			switch strings.ToLower(status) {
			case "online", "accessible", "active", "normal", "ready":
				return "1", nil // Estado OK
			case "offline", "inaccessible", "inactive", "error", "degraded":
				return "0", nil // Estado ERROR
			default:
				log.Printf("Estado desconocido de volumen: %s", status)
				return "2", nil // Estado DESCONOCIDO
			}
		}
	}

	// Si no se encuentra el volumen, retornar error
	log.Printf("Volumen no encontrado: %s", volumeName)
	return "0", nil
}

// GetNodeStatus obtiene el estado de un nodo específico por ID
// host: dirección IP o nombre de host del sistema FlashSystem
// user: nombre de usuario para autenticación
// pass: contraseña para autenticación
// nodeID: ID del nodo a consultar
// Retorna 1 para online, 0 para offline
func GetNodeStatus(host, user, pass, nodeID string) (string, error) {
	// Validar entrada para prevenir inyección de comandos
	if !isValidNumericInput(nodeID) {
		return "0", fmt.Errorf("argumento inválido para GetNodeStatus: node=%s", nodeID)
	}

	// Establecer conexión SSH
	client, err := ssh.EstablishSSHConnection(host, user, pass)
	if err != nil {
		return "0", fmt.Errorf("error estableciendo conexión SSH: %w", err)
	}
	defer func() {
		if closeErr := client.Close(); closeErr != nil {
			log.Printf("Advertencia: Error al cerrar conexión SSH: %v", closeErr)
		}
	}()

	// Ejecutar comando para obtener información de nodos
	cmd := "svcinfo lsnode -delim :"
	output, err := ssh.ExecuteCommand(client, cmd)
	if err != nil {
		return "0", fmt.Errorf("error ejecutando comando lsnode: %w", err)
	}
	// Parsear la salida del comando
	records := parser.ParseDelimitedOutput(output, ':')

	// Buscar el nodo específico en los registros
	for _, r := range records {
		if r["id"] == nodeID {
			// Convertir el estado a un valor numérico para Zabbix
			status := r["status"]
			switch strings.ToLower(status) {
			case "online", "active", "participating", "normal":
				return "1", nil // Estado OK
			case "offline", "inactive", "not_participating", "error", "degraded":
				return "0", nil // Estado ERROR
			default:
				log.Printf("Estado desconocido de nodo: %s", status)
				return "2", nil // Estado DESCONOCIDO
			}
		}
	}

	// Si no se encuentra el nodo, retornar error
	log.Printf("Nodo no encontrado: id=%s", nodeID)
	return "0", nil
}

// isValidNumericInput verifica si una cadena contiene solo caracteres numéricos
// Esta función es usada para validar entradas y prevenir inyección de comandos
func isValidNumericInput(input string) bool {
	// Permitir números enteros y negativos
	for _, r := range input {
		if !(r >= '0' && r <= '9' || r == '-') {
			return false
		}
	}
	return true
}

// convertToFloat convierte una cadena a número decimal
// Retorna el número y un error si la conversión falla
func convertToFloat(str string) (float64, error) {
	// Eliminar espacios en blanco y unidades comunes
	str = strings.TrimSpace(str)

	// Remover unidades comunes como 'GB', 'TB', etc. si existen
	// Este es un ejemplo simplificado, en producción podría ser más complejo
	if strings.HasSuffix(str, "GB") {
		str = strings.TrimSuffix(str, "GB")
		str = strings.TrimSpace(str)
	} else if strings.HasSuffix(str, "TB") {
		str = strings.TrimSuffix(str, "TB")
		str = strings.TrimSpace(str)
	} else if strings.HasSuffix(str, "MB") {
		str = strings.TrimSuffix(str, "MB")
		str = strings.TrimSpace(str)
	}

	return strconv.ParseFloat(str, 64)
}

// convertToInt convierte una cadena a número entero
// Retorna el número y un error si la conversión falla
func convertToInt(str string) (int, error) {
	f, err := convertToFloat(str)
	if err != nil {
		return 0, err
	}
	return int(f), nil
}

// GetClusterStatus obtiene el estado general del cluster
// host: dirección IP o nombre de host del sistema FlashSystem
// user: nombre de usuario para autenticación
// pass: contraseña para autenticación
// Retorna 1 para healthy, 0 para degraded/error
func GetClusterStatus(host, user, pass string) (string, error) {
	// Establecer conexión SSH
	client, err := ssh.EstablishSSHConnection(host, user, pass)
	if err != nil {
		return "0", fmt.Errorf("error estableciendo conexión SSH: %w", err)
	}
	defer func() {
		if closeErr := client.Close(); closeErr != nil {
			log.Printf("Advertencia: Error al cerrar conexión SSH: %v", closeErr)
		}
	}()

	// Ejecutar comando para obtener información del sistema
	cmd := "svcinfo lssystem -delim :"
	output, err := ssh.ExecuteCommand(client, cmd)
	if err != nil {
		return "0", fmt.Errorf("error ejecutando comando lssystem: %w", err)
	}

	// Parsear la salida del comando
	records := parser.ParseDelimitedOutput(output, ':')

	// Buscar el estado del sistema en los registros
	for _, r := range records {
		// El campo 'status' o 'cluster_ntp_sync' puede indicar el estado general
		status := r["status"]
		if status == "" {
			// Algunas versiones usan 'product_name' u otros campos para indicar estado
			productStatus := r["product_name"]
			if strings.Contains(strings.ToLower(productStatus), "error") {
				return "0", nil
			}
			// Si no hay campo status, asumimos healthy por defecto
			return "1", nil
		}

		switch strings.ToLower(status) {
		case "online", "active", "healthy", "normal", "operational":
			return "1", nil // Estado OK
		case "offline", "inactive", "degraded", "error", "problem":
			return "0", nil // Estado ERROR
		default:
			log.Printf("Estado desconocido del cluster: %s", status)
			return "2", nil // Estado DESCONOCIDO
		}
	}

	// Si no se encuentra información, retornar error
	log.Println("No se pudo obtener estado del cluster")
	return "0", nil
}
