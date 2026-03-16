// Package ssh proporciona funciones para conectar y comunicarse con sistemas IBM FlashSystem mediante SSH
// Este paquete encapsula toda la lógica de conexión SSH y ejecución de comandos remotos
package ssh

import (
	"bytes"
	"fmt"
	"time"

	"golang.org/x/crypto/ssh"
)

// EstablishSSHConnection crea una nueva conexión SSH al sistema de almacenamiento
// host: dirección IP o nombre de host del sistema FlashSystem
// user: nombre de usuario para autenticación
// pass: contraseña para autenticación
// Retorna un cliente SSH o un error si falla la conexión
func EstablishSSHConnection(host, user, pass string) (*ssh.Client, error) {
	// Configurar los parámetros de autenticación SSH
	config := &ssh.ClientConfig{
		// Usuario para autenticación
		User: user,
		// Método de autenticación: contraseña
		Auth: []ssh.AuthMethod{
			ssh.Password(pass),
		},
		// Tiempo máximo para establecer la conexión
		Timeout: 10 * time.Second,
		// Callback para verificar la clave del host remoto
		// ADVERTENCIA: En producción, debería usarse una verificación segura de clave de host
		// Para fines de desarrollo y pruebas, se ignora la verificación
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		// Identificador del cliente SSH
		ClientVersion: "SSH-2.0-FlashSystem-Zabbix-Monitor",
	}

	// Establecer la conexión SSH al puerto 22 del host especificado
	conn, err := ssh.Dial("tcp", host+":22", config)
	if err != nil {
		return nil, fmt.Errorf("error al conectar con el sistema FlashSystem: %w", err)
	}

	// Retornar el cliente SSH para futuras operaciones
	return conn, nil
}

// ExecuteCommand ejecuta un comando en el sistema FlashSystem remoto a través de SSH
// client: cliente SSH ya conectado
// command: comando SVC a ejecutar en el sistema remoto
// Retorna la salida del comando o un error si falla la ejecución
func ExecuteCommand(client *ssh.Client, command string) (string, error) {
	// Crear una nueva sesión SSH para ejecutar el comando
	session, err := client.NewSession()
	if err != nil {
		return "", fmt.Errorf("error al crear sesión SSH: %w", err)
	}
	// Asegurar que la sesión se cierra al finalizar la función
	defer func() {
		// Registrar cualquier error al cerrar la sesión
		if closeErr := session.Close(); closeErr != nil {
			// Usamos fmt para evitar recursión potencial en el sistema de logging
			fmt.Printf("Advertencia: Error al cerrar sesión SSH: %v\n", closeErr)
		}
	}()

	// Configurar un buffer para capturar la salida estándar del comando
	var outputBuffer bytes.Buffer
	session.Stdout = &outputBuffer

	// Configurar un buffer para capturar la salida de error del comando
	var errorBuffer bytes.Buffer
	session.Stderr = &errorBuffer

	// Establecer un tiempo límite para la ejecución del comando
	// Esto previene que el programa se quede colgado si el sistema remoto no responde
	done := make(chan error, 1)
	go func() {
		// Ejecutar el comando con el locale SVC configurado para asegurar formato consistente
		// SVC_CLI_LOCALE=C fuerza el uso del locale C (inglés) para la salida del comando
		fullCommand := "SVC_CLI_LOCALE=C " + command
		done <- session.Run(fullCommand)
	}()

	// Esperar la respuesta del comando o timeout
	select {
	case err := <-done:
		// Verificar si el comando falló
		if err != nil {
			// Si hay contenido en el buffer de error, incluirlo en el mensaje
			errorOutput := errorBuffer.String()
			if errorOutput != "" {
				return "", fmt.Errorf("comando '%s' falló con error: %s (stderr: %s): %w", command, err.Error(), errorOutput, err)
			}
			return "", fmt.Errorf("comando '%s' falló: %w", command, err)
		}
		// Comando ejecutado exitosamente, retornar la salida
		return outputBuffer.String(), nil
	case <-time.After(30 * time.Second):
		// Forzar cierre de sesión si se excede el tiempo límite
		session.Close()
		return "", fmt.Errorf("tiempo de espera excedido (30 segundos) al ejecutar comando: %s", command)
	}
}

// testConnection verifica si se puede establecer una conexión SSH con el sistema
// host: dirección IP o nombre de host del sistema FlashSystem
// user: nombre de usuario para autenticación
// pass: contraseña para autenticación
// Retorna nil si la conexión es exitosa, o un error si falla
func testConnection(host, user, pass string) error {
	// Establecer conexión SSH
	client, err := EstablishSSHConnection(host, user, pass)
	if err != nil {
		return fmt.Errorf("prueba de conexión fallida: %w", err)
	}
	defer func() {
		// Cerrar la conexión después de la prueba
		if closeErr := client.Close(); closeErr != nil {
			fmt.Printf("Advertencia: Error al cerrar conexión de prueba: %v\n", closeErr)
		}
	}()

	// Ejecutar un comando simple para verificar que la conexión funcione
	_, err = ExecuteCommand(client, "svcinfo lssystem -delim :")
	if err != nil {
		return fmt.Errorf("prueba de comando fallida: %w", err)
	}

	// Si llegamos aquí, la conexión es válida
	return nil
}

// validateCredentials verifica si las credenciales proporcionadas son válidas
// Intenta autenticarse con el sistema FlashSystem
// host: dirección IP o nombre de host del sistema FlashSystem
// user: nombre de usuario para autenticación
// pass: contraseña para autenticación
// Retorna true si las credenciales son válidas, false en caso contrario
func validateCredentials(host, user, pass string) (bool, error) {
	// Establecer conexión SSH
	client, err := EstablishSSHConnection(host, user, pass)
	if err != nil {
		// Si falla la conexión, probablemente las credenciales sean incorrectas
		return false, fmt.Errorf("credenciales inválidas o sistema no accesible: %w", err)
	}
	defer func() {
		// Cerrar la conexión después de la validación
		if closeErr := client.Close(); closeErr != nil {
			fmt.Printf("Advertencia: Error al cerrar conexión de validación: %v\n", closeErr)
		}
	}()

	// Intentar ejecutar un comando simple para confirmar que las credenciales funcionan
	// Usamos lssystem porque es un comando que debería estar disponible en todos los sistemas
	_, err = ExecuteCommand(client, "svcinfo lssystem -delim :")
	if err != nil {
		// Si el comando falla, las credenciales podrían ser válidas pero el usuario
		// podría no tener permisos para ejecutar el comando específico
		return false, fmt.Errorf("credenciales válidas pero sin permisos para comandos de monitoreo: %w", err)
	}

	// Si llegamos aquí, las credenciales son válidas y funcionan
	return true, nil
}

// getConnectionInfo obtiene información básica sobre el sistema conectado
// host: dirección IP o nombre de host del sistema FlashSystem
// user: nombre de usuario para autenticación
// pass: contraseña para autenticación
// Retorna un mapa con información del sistema o un error si falla
func getConnectionInfo(host, user, pass string) (map[string]string, error) {
	// Establecer conexión SSH
	client, err := EstablishSSHConnection(host, user, pass)
	if err != nil {
		return nil, fmt.Errorf("error al conectar para obtener información: %w", err)
	}
	defer func() {
		// Cerrar la conexión después de obtener la información
		if closeErr := client.Close(); closeErr != nil {
			fmt.Printf("Advertencia: Error al cerrar conexión de obtención de info: %v\n", closeErr)
		}
	}()

	// Ejecutar comando para obtener información del sistema
	output, err := ExecuteCommand(client, "svcinfo lssystem -delim :")
	if err != nil {
		return nil, fmt.Errorf("error al obtener información del sistema: %w", err)
	}

	// Parsear la salida para extraer información importante
	info := parseSystemInfo(output)
	return info, nil
}

// parseSystemInfo convierte la salida de lssystem en un mapa de valores
// output: salida del comando lssystem con delimitador :
// Retorna un mapa con los campos del sistema
func parseSystemInfo(output string) map[string]string {
	// Dividir la salida en líneas
	lines := splitLines(output)
	
	// Verificar que haya suficientes líneas (cabecera + datos)
	if len(lines) < 2 {
		return make(map[string]string)
	}

	// Extraer la cabecera (primera línea)
	headers := splitFields(lines[0], ":")
	
	// Extraer los valores (segunda línea)
	values := splitFields(lines[1], ":")
	
	// Crear el mapa de información
	info := make(map[string]string)
	
	// Emparejar cabeceras con valores
	for i, header := range headers {
		header = trimSpaces(header)
		if i < len(values) {
			info[header] = trimSpaces(values[i])
		} else {
			info[header] = ""
		}
	}

	return info
}

// splitLines divide una cadena en líneas, manejando diferentes tipos de saltos de línea
// text: texto a dividir en líneas
// Retorna un slice con cada línea
func splitLines(text string) []string {
	// Reemplazar \r\n con \n para normalizar los saltos de línea
	text = replaceAll(text, "\r\n", "\n")
	// Reemplazar \r con \n por si acaso
	text = replaceAll(text, "\r", "\n")
	
	// Dividir por \n
	lines := split(text, "\n")
	
	// Filtrar líneas vacías al principio y al final
	for len(lines) > 0 && lines[0] == "" {
		lines = lines[1:]
	}
	for len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	
	return lines
}

// splitFields divide una cadena por un delimitador específico
// text: texto a dividir
// delimiter: delimitador a usar
// Retorna un slice con los campos
func splitFields(text, delimiter string) []string {
	return split(text, delimiter)
}

// trimSpaces elimina espacios en blanco al principio y al final de una cadena
// s: cadena a recortar
// Retorna la cadena sin espacios en los extremos
func trimSpaces(s string) string {
	// Recorrer la cadena para encontrar el primer carácter que no sea espacio
	start := 0
	for start < len(s) && isSpace(s[start]) {
		start++
	}
	
	// Recorrer la cadena desde el final para encontrar el último carácter que no sea espacio
	end := len(s)
	for end > start && isSpace(s[end-1]) {
		end--
	}
	
	// Retornar la subcadena sin espacios
	return s[start:end]
}

// isSpace verifica si un byte es un carácter de espacio
// c: byte a verificar
// Retorna true si es un espacio, tabulador u otro carácter de espacio
func isSpace(c byte) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\r'
}

// split divide una cadena por un delimitador
// Esta es una implementación simple que no depende de paquetes adicionales
// text: texto a dividir
// delimiter: delimitador a usar
// Retorna un slice con las partes
func split(text, delimiter string) []string {
	var result []string
	delimiterLen := len(delimiter)
	
	if delimiterLen == 0 {
		// Si el delimitador es vacío, devolver cada caracter como un elemento
		for i := 0; i < len(text); i++ {
			result = append(result, string(text[i]))
		}
		return result
	}
	
	start := 0
	for i := 0; i <= len(text)-delimiterLen; i++ {
		if text[i:i+delimiterLen] == delimiter {
			result = append(result, text[start:i])
			i += delimiterLen - 1 // Ajustar índice para saltar el delimitador
			start = i + 1
		}
	}
	
	// Agregar la última parte
	result = append(result, text[start:])
	return result
}

// replaceAll reemplaza todas las ocurrencias de old por new en una cadena
// s: cadena original
// old: texto a reemplazar
// new: texto de reemplazo
// Retorna la cadena con los reemplazos
func replaceAll(s, old, new string) string {
	if old == "" {
		return s
	}
	
	// Contar cuántas veces aparece old
	count := 0
	oldLen := len(old)
	for i := 0; i <= len(s)-oldLen; i++ {
		if s[i:i+oldLen] == old {
			count++
			i += oldLen - 1
		}
	}
	
	if count == 0 {
		return s
	}
	
	// Crear la nueva cadena
	newLen := len(new)
	resultLen := len(s) + count*(newLen-oldLen)
	result := make([]byte, resultLen)
	
	j := 0
	start := 0
	for i := 0; i <= len(s)-oldLen; i++ {
		if s[i:i+oldLen] == old {
			// Copiar la parte antes del match
			copy(result[j:], s[start:i])
			j += i - start
			// Copiar el reemplazo
			copy(result[j:], new)
			j += newLen
			// Actualizar start
			start = i + oldLen
			// Ajustar i
			i += oldLen - 1
		}
	}
	// Copiar la parte final
	copy(result[j:], s[start:])
	
	return string(result[:j+len(s)-start])
}