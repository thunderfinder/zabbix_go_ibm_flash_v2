// Package parser proporciona funciones para analizar la salida de comandos CLI de IBM Spectrum Virtualize
// Este paquete convierte la salida de texto plano de comandos svcinfo en estructuras de datos Go
package parser

import (
	"strings"
	"unicode"
)

// Record representa un registro de datos de IBM Spectrum Virtualize
// Cada registro contiene campos clave-valor con información sobre un componente del sistema
type Record map[string]string

// ParseDelimitedOutput analiza la salida de un comando SVC delimitado por caracteres específicos
// output: la salida de texto del comando SVC
// delimiter: el carácter usado para separar campos (por defecto ':')
// Retorna una lista de registros con los datos parseados
func ParseDelimitedOutput(output string, delimiter rune) []Record {
	// Verificar que la salida no esté vacía
	if strings.TrimSpace(output) == "" {
		return nil
	}

	// Normalizar los caracteres de nueva línea para manejar diferentes sistemas
	output = strings.ReplaceAll(output, "\r\n", "\n")
	output = strings.ReplaceAll(output, "\r", "\n")

	// Dividir la salida en líneas
	lines := strings.Split(output, "\n")

	// Verificar que haya suficientes líneas (mínimo cabecera + datos)
	if len(lines) < 2 {
		return nil
	}

	// Extraer la cabecera (primera línea) que contiene los nombres de los campos
	headerLine := lines[0]
	headers := splitAndTrim(headerLine, delimiter)

	// Verificar que la cabecera contenga campos
	if len(headers) == 0 {
		return nil
	}

	// Preparar la lista para almacenar los registros parseados
	var records []Record

	// Procesar cada línea de datos (desde la segunda línea en adelante)
	for i := 1; i < len(lines); i++ {
		line := lines[i]
		
		// Saltar líneas vacías
		if strings.TrimSpace(line) == "" {
			continue
		}

		// Dividir la línea actual en valores usando el delimitador
		values := splitAndTrim(line, delimiter)

		// Crear un nuevo registro para esta línea
		record := make(Record)

		// Asociar cada valor con su campo correspondiente según la cabecera
		for j, header := range headers {
			// Verificar si hay un valor correspondiente para este campo
			if j < len(values) {
				// Asignar el valor al campo correspondiente
				record[header] = values[j]
			} else {
				// Si no hay valor, asignar cadena vacía
				record[header] = ""
			}
		}

		// Agregar el registro a la lista de resultados
		records = append(records, record)
	}

	// Retornar la lista de registros parseados
	return records
}

// splitAndTrim divide una cadena de texto por un delimitador y elimina espacios en blanco
// text: la cadena de texto a dividir
// delimiter: el carácter usado como delimitador
// Retorna una lista de cadenas con los campos resultantes, sin espacios en blanco
func splitAndTrim(text string, delimiter rune) []string {
	// Dividir la cadena por el delimitador
	parts := strings.FieldsFunc(text, func(c rune) bool {
		return c == delimiter
	})

	// Crear una lista para almacenar las partes limpias
	var trimmedParts []string

	// Recorrer cada parte y eliminar espacios en blanco al principio y al final
	for _, part := range parts {
		trimmedPart := strings.TrimSpace(part)
		trimmedParts = append(trimmedParts, trimmedPart)
	}

	// Retornar la lista de partes limpias
	return trimmedParts
}

// ParseSystemInfo convierte la salida de lssystem en un solo registro
// Esta función es útil cuando se espera solo un registro de información del sistema
// output: la salida del comando lssystem
// Retorna un único registro con la información del sistema o nil si falla
func ParseSystemInfo(output string) Record {
	// Parsear la salida como si fuera un comando delimitado
	records := ParseDelimitedOutput(output, ':')
	
	// Verificar que se haya obtenido al menos un registro
	if len(records) == 0 {
		return nil
	}
	
	// Retornar el primer (y único) registro de información del sistema
	return records[0]
}

// ParseNodeInfo convierte la salida de lsnode en registros// output: la salida del comando lsnode
// Retorna una lista de registros con la información de los nodos
func ParseNodeInfo(output string) []Record {
	// Parsear la salida usando el delimitador estándar ':'
	return ParseDelimitedOutput(output, ':')
}

// ParseDriveInfo convierte la salida de lsdrive en registros de unidades
// output: la salida del comando lsdrive
// Retorna una lista de registros con la información de las unidades
func ParseDriveInfo(output string) []Record {
	// Parsear la salida usando el delimitador estándar ':'
	return ParseDelimitedOutput(output, ':')
}

// ParseEnclosureInfo convierte la salida de lsenclosure en registros de carcasas
// output: la salida del comando lsenclosure
// Retorna una lista de registros con la información de las carcasas
func ParseEnclosureInfo(output string) []Record {
	// Parsear la salida usando el delimitador estándar ':'
	return ParseDelimitedOutput(output, ':')
}

//ierte la salida de lsmdiskgrp en registros de grupos de discos
// output: la salida del comando lsmdiskgrp
// Retorna una lista de registros con la información de los grupos de discos
func ParsePoolInfo(output string) []Record {
	// Parsear la salida usando el delimitador estándar ':'
	return ParseDelimitedOutput(output, ':')
}

// ParseVolumeInfo convierte la salida de lsvdisk en registros de volúmenes virtuales
// output: la salida del comando lsvdisk
// Retorna una lista de registros con la información de los volúmenes virtuales
func ParseVolumeInfo(output string) []Record {
	// Parsear la salida usando el delimitador estándar ':'
	return ParseDelimitedOutput(output, ':')
}

// ParsePortInfo convierte la salida de lsportfc en registros de puertos FC
// output: la salida del comando lsportfc
// Retorna una lista de registros con la información de los puertos Fibre Channel
func ParsePortInfo(output string) []Record {
	// Parsear la salida usando el delimitador estándar ':'
	return ParseDelimitedOutput(output, ':')
}

// ParseFlashCopyInfo convierte la salida de lsfcmap en registros de relaciones FlashCopy
// output: la salida del comando lsfcmap
// Retorna una lista de registros con la información de las relaciones FlashCopy
func ParseFlashCopyInfo(output string) []Record {
	// Parsear la salida usando el delimitador estándar ':'
	return ParseDelimitedOutput(output, ':')
}

// ParseReplicationInfo convierte la salida de lsrcrelationship en registros de replicación
// output: la salida del comando lsrcrelationship
// Retorna una lista de registros con la información de las relaciones de replicación
func ParseReplicationInfo(output string) []Record {
	// Parsear la salida usando el delimitador estándar ':'
	return ParseDelimitedOutput(output, ':')
}

// ParsePerformanceInfo convierte la salida de lssystemstats en registros de rendimiento
// output: la salida del comando lssystemstats
// Retorna una lista de registros con la información de rendimiento del sistema
func ParsePerformanceInfo(output string) []Record {
	// Parsear la salida usando el delimitador estándar ':'
	return ParseDelimitedOutput(output, ':')
}

// CleanValue elimina caracteres especiales y espacios innecesarios de un valor
// value: el valor a limpiar
// Retorna el valor limpio y listo para su procesamiento
func CleanValue(value string) string {
	// Eliminar espacios en blanco al principio y al final
	value = strings.TrimSpace(value)
	
	// Eliminar caracteres de control que podrían causar problemas
	var cleaned strings.Builder
	for _, r := range value {
		if !unicode.IsControl(r) {
			cleaned.WriteRune(r)
		}
	}
	
	return cleaned.String()
}

// IsValidRecord verifica si un registro contiene datos válidos
// Un registro es válido si tiene al menos un campo con valor no vacío
// record: el registro a validar
// Retorna true si el registro es válido, false en caso contrario
func IsValidRecord(record Record) bool {
	// Verificar que el registro no sea nil
	if record == nil {
		return false
	}
	
	// Verificar que al menos un campo tenga un valor no vacío
	for _, value := range record {
		if value != "" {
			return true
		}
	}
	
	// Si todos los campos están vacíos, el registro no es válido
	return false
}

// FilterValidRecords elimina los registros que no contienen datos válidos
// records: la lista de registros a filtrar
// Retorna una nueva lista solo con los registros válidos
func FilterValidRecords(records []Record) []Record {
	var validRecords []Record
	
	// Recorrer cada registro y verificar si es válido
	for _, record := range records {
		if IsValidRecord(record) {
			validRecords = append(validRecords, record)
		}
	}
	
	return validRecords
}

// GetFieldValue busca un valor específico en un registro por su campo
// record: el registro donde buscar
// fieldName: el nombre del campo a buscar
// defaultValue: el valor a retornar si el campo no se encuentra
// Retorna el valor del campo o el valor por defecto si no se encuentra
func GetFieldValue(record Record, fieldName, defaultValue string) string {
	// Verificar que el registro no sea nil
	if record == nil {
		return defaultValue
	}
	
	// Buscar el campo en el registro
	value, exists := record[fieldName]
	if !exists {
		return defaultValue
	}
	
	// Retornar el valor encontrado
	return value
}

// HasField verifica si un registro contiene un campo específico
// record: el registro donde buscar
// fieldName: el nombre del campo a verificar
// Retorna true si el campo existe, false en caso contrario
func HasField(record Record, fieldName string) bool {
	// Verificar que el registro no sea nil
	if record == nil {
		return false
	}
	
	// Verificar si el campo existe en el registro
	_, exists := record[fieldName]
	return exists
}

// GetUniqueFieldValues obtiene todos los valores únicos de un campo específico
// records: la lista de registros donde buscar
// fieldName: el nombre del campo a extraer
// Retorna una lista con los valores únicos del campo
func GetUniqueFieldValues(records []Record, fieldName string) []string {
	// Mapa para almacenar valores únicos
	uniqueValues := make(map[string]bool)
	var result []string
	
	// Recorrer cada registro
	for _, record := range records {
		// Obtener el valor del campo
		value := GetFieldValue(record, fieldName, "")
		
		// Si el valor no está en el mapa, agregarlo
		if !uniqueValues[value] {
			uniqueValues[value] = true
			result = append(result, value)
		}
	}
	
	return result
}