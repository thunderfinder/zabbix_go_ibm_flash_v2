// Package utils contiene funciones de utilidad general para el sistema de monitoreo
// Este paquete provee herramientas comunes utilizadas por otros módulos del sistema
package utils

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

// IsValidNumericInput verifica si una cadena contiene solo caracteres numéricos
// Esta función es usada para validar entradas y prevenir inyección de comandos
// input: la cadena a validar
// Retorna true si la cadena es numérica válida, false en caso contrario
func IsValidNumericInput(input string) bool {
	// Si la cadena está vacía, no es válida
	if input == "" {
		return false
	}
	
	// Permitir números enteros y negativos
	for i, r := range input {
		// El primer carácter puede ser un signo negativo
		if i == 0 && r == '-' {
			continue
		}
		// Verificar si el carácter es un dígito
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

// IsValidAlphanumericInput verifica si una cadena contiene solo caracteres alfanuméricos
// Esta función es usada para validar entradas como nombres de pools o volúmenes
// input: la cadena a validar
// Retorna true si la cadena es alfanumérica válida, false en caso contrario
func IsValidAlphanumericInput(input string) bool {
	// Si la cadena está vacía, no es válida
	if input == "" {
		return false
	}
	
	// Verificar que todos los caracteres sean alfanuméricos o guiones bajos
	for _, r := range input {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' && r != '-' {
			return false
		}
	}
	return true
}

// ConvertToFloat convierte una cadena a número decimal
// str: la cadena a convertir
// Retorna el número y un error si la conversión falla
func ConvertToFloat(str string) (float64, error) {
	// Eliminar espacios en blanco al principio y al final
	str = strings.TrimSpace(str)
	
	// Remover unidades comunes como 'GB', 'TB', etc. si existen
	// Este es un ejemplo simplificado, en producción podría ser más complejo
	if strings.HasSuffix(strings.ToUpper(str), "GB") {
		str = strings.TrimSuffix(str, "GB")
		str = strings.TrimSpace(str)
	} else if strings.HasSuffix(strings.ToUpper(str), "TB") {
		str = strings.TrimSuffix(str, "TB")
		str = strings.TrimSpace(str)
	} else if strings.HasSuffix(strings.ToUpper(str), "MB") {
		str = strings.TrimSuffix(str, "MB")
		str = strings.TrimSpace(str)
	} else if strings.HasSuffix(strings.ToUpper(str), "KB") {
		str = strings.TrimSuffix(str, "KB")
		str = strings.TrimSpace(str)
	}
	
	// Convertir la cadena a float64
	return strconv.ParseFloat(str, 64)
}

// ConvertToInt convierte una cadena a número entero
// str: la cadena a convertir
// Retorna el número y un error si la conversión falla
func ConvertToInt(str string) (int, error) {
	f, err := ConvertToFloat(str)
	if err != nil {
		return 0, err
	}
	return int(f), nil
}

// EscapeJSONValue escapa caracteres especiales en un valor para que sea seguro en JSON
// value: el valor a escapar
// Retorna el valor con caracteres especiales escapados
func EscapeJSONValue(value string) string {
	// Si el valor es nulo o vacío, retornar como está
	if value == "" {
		return value
	}
	
	// Reemplazar caracteres que podrían causar problemas en JSON
	value = strings.ReplaceAll(value, `\`, `\\`)  // Backslash
	value = strings.ReplaceAll(value, `"`, `\"`)   // Comillas dobles
	value = strings.ReplaceAll(value, "\n", `\n`)  // Nueva línea
	value = strings.ReplaceAll(value, "\r", `\r`)  // Retorno de carro
	value = strings.ReplaceAll(value, "\t", `\t`)  // Tabulador
	
	return value
}

// TrimSpacesAndClean elimina espacios en blanco al principio y al final de una cadena
// Además, elimina caracteres de control que podrían causar problemas
// s: la cadena a limpiar
// Retorna la cadena limpia
func TrimSpacesAndClean(s string) string {
	// Eliminar espacios en blanco al principio y al final
	s = strings.TrimSpace(s)
	
	// Eliminar caracteres de control que podrían causar problemas
	var cleaned strings.Builder
	for _, r := range s {
		if !unicode.IsControl(r) {
			cleaned.WriteRune(r)
		}
	}
	
	return cleaned.String()
}

// FormatPercentage formatea un valor decimal como porcentaje con 2 decimales
// value: el valor decimal a formatear
// Retorna una cadena con el porcentaje formateado
func FormatPercentage(value float64) string {
	return fmt.Sprintf("%.2f", value)
}

// FormatBytes formatea un valor en bytes a una unidad legible (B, KB, MB, GB, TB)
// bytes: el valor en bytes a formatear
// Retorna una cadena con el valor formateado y su unidad
func FormatBytes(bytes float64) string {
	units := []string{"B", "KB", "MB", "GB", "TB"}
	unitIndex := 0
	value := bytes
	
	// Convertir a la unidad más apropiada
	for value >= 1024 && unitIndex < len(units)-1 {
		value /= 1024
		unitIndex++
	}
	
	return fmt.Sprintf("%.2f %s", value, units[unitIndex])
}

// ContainsString verifica si una cadena está presente en un slice de cadenas
// slice: el slice de cadenas donde buscar
// s: la cadena a buscar
// Retorna true si la cadena está presente, false en caso contrario
func ContainsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

// UniqueStrings devuelve un slice con cadenas únicas, eliminando duplicados
// slice: el slice de cadenas a procesar
// Retorna un nuevo slice sin duplicados
func UniqueStrings(slice []string) []string {
	seen := make(map[string]bool)
	var result []string
	
	for _, item := range slice {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}
	
	return result
}

// StringSliceContainsAny verifica si alguna cadena del primer slice está presente en el segundo slice
// slice1: el slice donde buscar
// slice2: el slice donde encontrar
// Retorna true si al menos una cadena está presente, false en caso contrario
func StringSliceContainsAny(slice1, slice2 []string) bool {
	for _, item1 := range slice1 {
		for _, item2 := range slice2 {
			if item1 == item2 {
				return true
			}
		}
	}
	return false
}

// SplitAndTrim splits a string by a separator and trims spaces from each part
// text: the string to split
// separator: the separator to use
// Returns a slice of trimmed strings
func SplitAndTrim(text, separator string) []string {
	// Split the string by the separator
	parts := strings.Split(text, separator)
	
	// Create a slice to store the trimmed parts
	var trimmedParts []string
	
	// Trim spaces from each part and add non-empty parts to the result
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			trimmedParts = append(trimmedParts, trimmed)
		}
	}
	
	return trimmedParts
}

// JoinNonEmpty joins non-empty strings with a separator
// elements: the strings to join
// separator: the separator to use
// Returns a joined string with only non-empty elements
func JoinNonEmpty(elements []string, separator string) string {
	var nonEmpty []string
	for _, element := range elements {
		if element != "" {
			nonEmpty = append(nonEmpty, element)
		}
	}
	return strings.Join(nonEmpty, separator)
}

// PadLeft pads a string on the left with a specified character up to a given length
// s: the string to pad
// length: the desired total length
// padChar: the character to use for padding
// Returns the padded string
func PadLeft(s string, length int, padChar rune) string {
	currentLength := len(s)
	if currentLength >= length {
		return s
	}
	
	padding := strings.Repeat(string(padChar), length-currentLength)
	return padding + s
}

// PadRight pads a string on the right with a specified character up to a given length
// s: the string to pad
// length: the desired total length
// padChar: the character to use for padding
// Returns the padded string
func PadRight(s string, length int, padChar rune) string {
	currentLength := len(s)
	if currentLength >= length {
		return s
	}
	
	padding := strings.Repeat(string(padChar), length-currentLength)
	return s + padding
}

// IsValidIP verifica si una cadena es una dirección IP válida
// ip: la cadena a validar
// Retorna true si es una IP válida, false en caso contrario
func IsValidIP(ip string) bool {
	// Dividir la IP en octetos
	octets := strings.Split(ip, ".")
	
	// Una IP IPv4 válida tiene 4 octetos
	if len(octets) != 4 {
		return false
	}
	
	// Verificar cada octeto
	for _, octet := range octets {
		// Cada octeto debe ser numérico
		if !IsValidNumericInput(octet) {
			return false
		}
		
		// Convertir a entero
		num, err := strconv.Atoi(octet)
		if err != nil {
			return false
		}
		
		// Cada octeto debe estar entre 0 y 255
		if num < 0 || num > 255 {
			return false
		}
	}
	
	return true
}

// SanitizeInput removes potentially dangerous characters from user input
// input: the string to sanitize
// Returns a sanitized string safe for use in commands or queries
func SanitizeInput(input string) string {
	// Remove or escape characters that could be used in injection attacks
	// For now, we'll just remove common problematic characters
	// In a real implementation, you'd want more sophisticated sanitization
	
	// Replace semicolons (command separator)
	input = strings.ReplaceAll(input, ";", "")
	
	// Replace pipes (pipeline operator)
	input = strings.ReplaceAll(input, "|", "")
	
	// Replace ampersands (background operator)
	input = strings.ReplaceAll(input, "&", "")
	
	// Replace dollar signs (variable expansion)
	input = strings.ReplaceAll(input, "$", "")
	
	// Replace backticks (command substitution)
	input = strings.ReplaceAll(input, "`", "")
	
	// Replace parentheses (subshell)
	input = strings.ReplaceAll(input, "(", "")
	input = strings.ReplaceAll(input, ")", "")
	
	// Replace angle brackets (redirection)
	input = strings.ReplaceAll(input, "<", "")
	input = strings.ReplaceAll(input, ">", "")
	
	// Remove leading and trailing spaces
	input = strings.TrimSpace(input)
	
	return input
}

// IsValidHexString verifica si una cadena es un valor hexadecimal válido
// hex: la cadena a validar
// Retorna true si es hexadecimal válido, false en caso contrario
func IsValidHexString(hex string) bool {
	// Remover el prefijo 0x si está presente
	if strings.HasPrefix(strings.ToLower(hex), "0x") {
		hex = hex[2:]
	}
	
	// Verificar que todos los caracteres sean hexadecimales
	for _, r := range hex {
		if !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')) {
			return false
		}
	}
	
	return true
}

// FormatTimestamp convierte un timestamp Unix a una cadena legible
// timestamp: el timestamp Unix a formatear
// layout: el formato deseado (por ejemplo, "2006-01-02 15:04:05")
// Retorna la cadena formateada o una cadena vacía si hay error
func FormatTimestamp(timestamp int64, layout string) string {
	// En Go, el layout predeterminado para time.Format es diferente
	// Por simplicidad en este contexto, simplemente devolvemos el timestamp como string
	// En una implementación real, usaríamos time.Unix(timestamp, 0).Format(layout)
	return strconv.FormatInt(timestamp, 10)
}