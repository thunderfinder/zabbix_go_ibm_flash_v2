// Package main es el punto de entrada para el programa de monitoreo de IBM FlashSystem
// Este programa interactúa con el sistema de almacenamiento IBM FlashSystem a través de SSH
// y proporciona métricas para el sistema de monitoreo Zabbix.
package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"flashsystem_zabbix/internal/discovery"
	"flashsystem_zabbix/internal/monitor"
	"flashsystem_zabbix/internal/parser"
	"flashsystem_zabbix/internal/ssh"
	"flashsystem_zabbix/internal/utils"
)

// init configura el sistema de logging para registrar eventos importantes
// durante la ejecución del programa. Los logs se escriben en stderr para facilitar
// la depuración en entornos de producción.
func init() {
	// Configurar el logger para incluir fecha y hora
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	// Registrar el inicio del programa
	log.Println("Iniciando monitor IBM FlashSystem")
}

// main es la función principal del programa
// Procesa los argumentos de línea de comandos y ejecuta el comando solicitado
func main() {
	// Validar que se hayan proporcionado suf
	// Se requieren al menos 4 argumentos: host, usuario, contraseña y comando
	if len(os.Args) < 5 {
		log.Println("Número insuficiente de argumentos. Se requieren al menos 4: host, usuario, contraseña, comando")
		fmt.Println("0") // Valor de error para Zabbix
		os.Exit(1)
	}

	// Extraer los argumentos de línea de comandos
	host := os.Args[1]     // Dirección IP o nombre de host del sistema FlashSystem
	user := os.Args[2]     // Nombre de usuario para autenticación SSH
	pass := os.Args[3]     // Contraseña para autenticación SSH
	command := os.Args[4]  // Comando de monitoreo a ejecutar

	// Validar que los argumentos no estén vacíos
	if host == "" {
		log.Println("El argumento host no puede estar vacío")
		fmt.Println("0")
		os.Exit(1)
	}
	if user == "" {
		log.Println("El argumento usuario no puede estar vacío")
		fmt.Println("0")
		os.Exit(1)
	}
	if pass == "" {
		log.Println("El argumento contraseña no puede estar vacío")
		fmt.Println("0")
		os.Exit(1)
	}
	if command == "" {
		log.Println("El argumento comando no puede estar vacío")
		fmt.Println("0")
		os.Exit(1)
	}

	// Validar que el comando esté en la lista de comandos permitidos
	// Esto previene la ejecución de comandos no autorizados
	if !isValidCommand(command) {
		log.Printf("Comando no permitido: %s", command)
		fmt.Println("0")
		os.Exit(1)
	}

	// Ejecutar el comando solicitado
	// Cada comando tiene su propia lógica de procesamiento
	switch command {
	case "discoverdrives":
		// Descubre unidades de disco disponibles en el sistema
		// Este comando es útil para crear reglas de descubrimiento dinámico en Zabbix
		executeDiscoverDrives(host, user, pass)

	case "discoverenclosures":
		// Descubre carcasas (enclosures) disponibles en el sistema
		// Similar al descubrimiento de unidades, pero para carcasas
		executeDiscoverEnclosures(host, user, pass)

	case "discoverpools":
		// Descubre grupos de discos disponibles en el sistema
		executeDiscoverPools(host, user, pass)

	case "discovervolumes":
		// Descubre volúmenes virtuales disponibles en el sistema
		executeDiscoverVolumes(host, user, pass)

	case "discovernodes":
		// Descubre nodos disponibles en el sistema
		executeDiscoverNodes(host, user, pass)

	case "discoverfcports":
		// Descubre puertos FC disponibles en el sistema
		executeDiscoverFCPorts(host, user, pass)

	case "discoverflashcopy":
		// Descubre relaciones FlashCopy disponibles en el sistema
		executeDiscoverFlashCopy(host, user, pass)

	case "discoverreplication":
		// Descubre relaciones de replicación disponibles en el sistema
		executeDiscoverReplication(host, user, pass)

	case "getdrivestatus":
		// Obtiene el estado de una unidad específica identificada por enclosure y slot
		// Requiere dos argumentos adicionales: enclosure ID y slot ID
		if len(os.Args) < 7 {
			log.Println("getdrivestatus requiere 2 argumentos adicionales: enclosure_id y slot_id")
			fmt.Println("0")
			return
		}
		enclosureID := os.Args[5]
		slotID := os.Args[6]
		executeGetDriveStatus(host, user, pass, enclosureID, slotID)

	case "getenclosurestatus":
		// Obtiene el estado de una carcasa específica
		// Requiere un argumento adicional: el ID de la carcasa
		if len(os.Args) < 6 {
			log.Println("getenclosurestatus requiere 1 argumento adicional: enclosure_id")
			fmt.Println("0")
			return
		}
		enclosureID := os.Args[5]
		executeGetEnclosureStatus(host, user, pass, enclosureID)

	case "getbatterystatus":
		// Obtiene el estado de la batería de una carcasa específica
		// Requiere un argumento adicional: el ID de la carcasa
		if len(os.Args) < 6 {
			log.Println("getbatterystatus requiere 1 argumento adicional: enclosure_id")
			fmt.Println("0")
			return
		}
		enclosureID := os.Args[5]
		executeGetBatteryStatus(host, user, pass, enclosureID)

	case "getpoolusage":
		// Obtiene el porcentaje de uso de un grupo de discos (pool)
		// Requiere un argumento adicional: el nombre del pool
		if len(os.Args) < 6 {
			log.Println("getpoolusage requiere 1 argumento adicional: pool_name")
			fmt.Println("0")
			return
		}
		poolName := os.Args[5]
		executeGetPoolUsage(host, user, pass, poolName)

	case "getiops":
		// Obtiene las operaciones de entrada/salida por segundo (IOPS) del sistema
		// Este comando no requiere argumentos adicionales
		executeGetIOPS(host, user, pass)

	case "getvolumestatus":
		// Obtiene el estado de un volumen específico
		if len(os.Args) < 6 {
			log.Println("getvolumestatus requiere 1 argumento adicional: volume_name")
			fmt.Println("0")
			return
		}
		volumeName := os.Args[5]
		executeGetVolumeStatus(host, user, pass, volumeName)

	case "getnodestatus":
		// Obtiene el estado de un nodo específico
		if len(os.Args) < 6 {
			log.Println("getnodestatus requiere 1 argumento adicional: node_id")
			fmt.Println("0")
			return
		}
		nodeID := os.Args[5]
		executeGetNodeStatus(host, user, pass, nodeID)

	case "getclusterstatus":
		// Obtiene el estado general del cluster
		executeGetClusterStatus(host, user, pass)

	default:
		// Si el comando no está implementado, registrar el error y salir
		log.Printf("Comando no implementado: %s", command)
		fmt.Println("0")
		os.Exit(1)
	}
}

// isValidCommand verifica si un comando está en la lista de comandos permitidos
// Esta función es crucial para la seguridad del sistema ya que previene
// la ejecución de comandos no autorizados
func isValidCommand(command string) bool {
	// Definir la lista de comandos permitidos
	allowedCommands := map[string]bool{
		"discoverdrives":      true,
		"discoverenclosures":  true,
		"discoverpools":       true,
		"discovervolumes":     true,
		"discovernodes":       true,
		"discoverfcports":     true,
		"discoverflashcopy":   true,
		"discoverreplication": true,
		"getdrivestatus":      true,
		"getenclosurestatus":  true,
		"getbatterystatus":    true,
		"getpoolusage":        true,
		"getiops":             true,
		"getvolumestatus":     true,
		"getnodestatus":       true,
		"getclusterstatus":    true,
	}

	// Verificar si el comando está en la lista de permitidos
	_, exists := allowedCommands[command]
	return exists
}

// executeDiscoverDrives ejecuta la lógica para descubrir unidades de disco
// Retorna un JSON con la información de las unidades encontradas
func executeDiscoverDrives(host, user, pass string) {
	result, err := discovery.DiscoverDrives(host, user, pass)
	if err != nil {
		log.Printf("Error al descubrir unidades: %v", err)
		fmt.Println("{}")
		return
	}
	fmt.Println(result)
}

// executeDiscoverEnclosures ejecuta la lógica para descubrir carcasas
// Retorna un JSON con la información de las carcasas encontradas
func executeDiscoverEnclosures(host, user, pass string) {
	result, err := discovery.DiscoverEnclosures(host, user, pass)
	if err != nil {
		log.Printf("Error al descubrir carcasas: %v", err)
		fmt.Println("{}")
		return
	}
	fmt.Println(result)
}

// executeDiscoverPools ejecuta la lógica para descubrir grupos de discos
// Retorna un JSON con la información de los grupos de discos encontrados
func executeDiscoverPools(host, user, pass string) {
	result, err := discovery.DiscoverPools(host, user, pass)
	if err != nil {
		log.Printf("Error al descubrir grupos de discos: %v", err)
		fmt.Println("{}")
		return
	}
	fmt.Println(result)
}

// executeDiscoverVolumes ejecuta la lógica para descubrir volúmenes virtuales
// Retorna un JSON con la información de los volúmenes encontrados
func executeDiscoverVolumes(host, user, pass string) {
	result, err := discovery.DiscoverVolumes(host, user, pass)
	if err != nil {
		log.Printf("Error al descubrir volúmenes: %v", err)
		fmt.Println("{}")
		return
	}
	fmt.Println(result)
}

// executeDiscoverNodes ejecuta la lógica para descubrir nodos
// Retorna un JSON con la información de los nodos encontrados
func executeDiscoverNodes(host, user, pass string) {
	result, err := discovery.DiscoverNodes(host, user, pass)
	if err != nil {
		log.Printf("Error al descubrir nodos: %v", err)
		fmt.Println("{}")
		return
	}
	fmt.Println(result)
}

// executeDiscoverFCPorts ejecuta la lógica para descubrir puertos FC
// Retorna un JSON con la información de los puertos FC encontrados
func executeDiscoverFCPorts(host, user, pass string) {
	result, err := discovery.DiscoverFCPorts(host, user, pass)
	if err != nil {
		log.Printf("Error al descubrir puertos FC: %v", err)
		fmt.Println("{}")
		return
	}
	fmt.Println(result)
}

// executeDiscoverFlashCopy ejecuta la lógica para descubrir relaciones FlashCopy
// Retorna un JSON con la información de las relaciones FlashCopy encontradas
func executeDiscoverFlashCopy(host, user, pass string) {
	result, err := discovery.DiscoverFlashCopyRelationships(host, user, pass)
	if err != nil {
		log.Printf("Error al descubrir relaciones FlashCopy: %v", err)
		fmt.Println("{}")
		return
	}
	fmt.Println(result)
}

// executeDiscoverReplication ejecuta la lógica para descubrir relaciones de replicación
// Retorna un JSON con la información de las relaciones de replicación encontradas
func executeDiscoverReplication(host, user, pass string) {
	result, err := discovery.DiscoverReplicationRelationships(host, user, pass)
	if err != nil {
		log.Printf("Error al descubrir relaciones de replicación: %v", err)
		fmt.Println("{}")
		return
	}
	fmt.Println(result)
}

// executeGetDriveStatus ejecuta la lógica para obtener el estado de una unidad específica
// Retorna un valor numérico representando el estado de la unidad
func executeGetDriveStatus(host, user, pass, enclosureID, slotID string) {
	status, err := monitor.GetDriveStatus(host, user, pass, enclosureID, slotID)
	if err != nil {
		log.Printf("Error al obtener estado de unidad: %v", err)
		fmt.Println("0")
		return
	}
	fmt.Println(status)
}

// executeGetEnclosureStatus ejecuta la lógica para obtener el estado de una carcasa específica
// Retorna un valor numérico representando el estado de la carcasa
func executeGetEnclosureStatus(host, user, pass, enclosureID string) {
	status, err := monitor.GetEnclosureStatus(host, user, pass, enclosureID)
	if err != nil {
		log.Printf("Error al obtener estado de carcasa: %v", err)
		fmt.Println("0")
		return
	}
	fmt.Println(status)
}

// executeGetBatteryStatus ejecuta la lógica para obtener el estado de la batería de una carcasa específica
// Retorna un valor numérico representando el estado de la batería
func executeGetBatteryStatus(host, user, pass, enclosureID string) {
	status, err := monitor.GetBatteryStatus(host, user, pass, enclosureID)
	if err != nil {
		log.Printf("Error al obtener estado de batería: %v", err)
		fmt.Println("0")
		return
	}
	fmt.Println(status)
}

// executeGetPoolUsage ejecuta la lógica para obtener el porcentaje de uso de un grupo de discos
// Retorna un valor decimal representando el porcentaje de uso
func executeGetPoolUsage(host, user, pass, poolName string) {
	usage, err := monitor.GetPoolUsage(host, user, pass, poolName)
	if err != nil {
		log.Printf("Error al obtener uso de grupo de discos: %v", err)
		fmt.Println("0")
		return
	}
	fmt.Printf("%.2f\n", usage)
}

// executeGetIOPS ejecuta la lógica para obtener las operaciones de entrada/salida por segundo
// Retorna un valor entero representando las IOPS del sistema
func executeGetIOPS(host, user, pass string) {
	iops, err := monitor.GetTotalIOPS(host, user, pass)
	if err != nil {
		log.Printf("Error al obtener IOPS: %v", err)
		fmt.Println("0")
		return
	}
	fmt.Println(iops)
}

// executeGetVolumeStatus ejecuta la lógica para obtener el estado de un volumen específico
// Retorna un valor numérico representando el estado del volumen
func executeGetVolumeStatus(host, user, pass, volumeName string) {
	status, err := monitor.GetVolumeStatus(host, user, pass, volumeName)
	if err != nil {
		log.Printf("Error al obtener estado de volumen: %v", err)
		fmt.Println("0")
		return
	}
	fmt.Println(status)
}

// executeGetNodeStatus ejecuta la lógica para obtener el estado de un nodo específico
// Retorna un valor numérico representando el estado del nodo
func executeGetNodeStatus(host, user, pass, nodeID string) {
	status, err := monitor.GetNodeStatus(host, user, pass, nodeID)
	if err != nil {
		log.Printf("Error al obtener estado de nodo: %v", err)
		fmt.Println("0")
		return
	}
	fmt.Println(status)
}

// executeGetClusterStatus ejecuta la lógica para obtener el estado general del cluster
// Retorna un valor numérico representando el estado del cluster
func executeGetClusterStatus(host, user, pass string) {
	status, err := monitor.GetClusterStatus(host, user, pass)
	if err != nil {
		log.Printf("Error al obtener estado del cluster: %v", err)
		fmt.Println("0")
		return
	}
	fmt.Println(status)
}

// runSSHCommandWithRetry ejecuta un comando SSH con lógica de reintento
// Intenta ejecutar el comando hasta 3 veces si falla
func runSSHCommandWithRetry(host, user, pass, cmd string) (string, error) {
	var output string
	var err error
	
	// Definir el número máximo de intentos
	maxRetries := 3
	
	// Intentar ejecutar el comando hasta maxRetries veces
	for attempt := 0; attempt < maxRetries; attempt++ {
		// Registrar intento
		log.Printf("Intento %d/%d para ejecutar comando: %s", attempt+1, maxRetries, cmd)
		
		// Ejecutar el comando SSH
		output, err = runSSHCommand(host, user, pass, cmd)
		if err == nil {
			// Si no hubo error, retornar el resultado
			return output, nil
		}
		
		// Registrar error
		log.Printf("Intento %d fallido: %v", attempt+1, err)
		
		// Esperar antes de reintentar (excepto en el último intento)
		if attempt < maxRetries-1 {
			waitTime := time.Duration(attempt+1) * time.Second
			log.Printf("Esperando %v antes de reintentar...", waitTime)
			time.Sleep(waitTime)
		}
	}
	
	// Si se agotaron los intentos, retornar el último error
	return "", fmt.Errorf("comando '%s' falló después de %d intentos: %w", cmd, maxRetries, err)
}

// runSSHCommand ejecuta un comando SSH en el sistema remoto
// Retorna la salida del comando o un error si falla
func runSSHCommand(host, user, pass, cmd string) (string, error) {
	// Establecer una nueva conexión SSH para cada comando
	// Esto evita problemas con conexiones persistentes en entornos de producción
	client, err := ssh.EstablishSSHConnection(host, user, pass)
	if err != nil {
		return "", fmt.Errorf("error estableciendo conexión SSH: %w", err)
	}
	defer func() {
		if closeErr := client.Close(); closeErr != nil {
			log.Printf("Error cerrando conexión SSH: %v", closeErr)
		}
	}()

	// Crear una nueva sesión SSH
	session, err := client.NewSession()
	if err != nil {
		return "", fmt.Errorf("error creando sesión SSH: %w", err)
	}
	defer func() {
		if closeErr := session.Close(); closeErr != nil {
			log.Printf("Error cerrando sesión SSH: %v", closeErr)
		}
	}()

	// Configurar el buffer para capturar la salida
	var outputBuffer strings.Builder
	session.Stdout = &outputBuffer

	// Establecer un timeout para prevenir que el comando se quede colgado
	done := make(chan error, 1)
	go func() {
		// Ejecutar el comando con el locale SVC configurado
		// SVC_CLI_LOCALE=C asegura que la salida esté en formato consistente
		fullCmd := "SVC_CLI_LOCALE=C " + cmd
		done <- session.Run(fullCmd)
	}()

	// Esperar la respuesta del comando o timeout
	select {
	case err := <-done:
		if err != nil {
			return "", fmt.Errorf("error ejecutando comando '%s': %w", cmd, err)
		}
		return outputBuffer.String(), nil
	case <-time.After(30 * time.Second):
		// Forzar cierre de sesión si se excede el timeout
		session.Close()
		return "", fmt.Errorf("timeout después de 30 segundos ejecutando comando: %s", cmd)
	}
}

// parseDelimitedOutput parsea la salida de un comando SVC separado por dos puntos
// Retorna una lista de mapas donde cada mapa representa una fila de datos
func parseDelimitedOutput(output string) []map[string]string {
	// Dividir la salida en líneas
	lines := strings.Split(strings.TrimSpace(output), "\n")
	
	// Verificar que haya suficientes líneas (cabecera + datos)
	if len(lines) < 2 {
		return nil
	}

	// Extraer la cabecera (primera línea)
	headers := strings.Split(lines[0], ":")
	
	// Preparar la lista para almacenar los registros
	var results []map[string]string

	// Procesar cada línea de datos (desde la segunda línea)
	for _, line := range lines[1:] {
		// Saltar líneas vacías
		if strings.TrimSpace(line) == "" {
			continue
		}
		
		// Dividir la línea en valores usando el separador ':'
		values := strings.Split(line, ":")
		
		// Crear un mapa para almacenar los pares campo:valor
		row := make(map[string]string)
		
		// Asociar cada valor con su respectivo encabezado
		for i, h := range headers {
			h = strings.TrimSpace(h) // Limpiar espacios en blanco
			if i < len(values) {
				row[h] = strings.TrimSpace(values[i]) // Limpiar espacios en blanco
			} else {
				row[h] = "" // Si no hay valor, asignar cadena vacía
			}
		}

		// Agregar el registro a la lista de resultados
		results = append(results, row)
	}

	// Devolver la lista de registros
	return results
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
	}
	
	return fmt.ParseFloat(str, 64)
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