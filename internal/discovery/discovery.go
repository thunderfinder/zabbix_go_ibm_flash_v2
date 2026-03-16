// Package discovery implementa funciones de descubrimiento automático para Zabbix
// Este paquete genera JSON en el formato requerido por Zabbix para reglas de descubrimiento dinámico
package discovery

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"zabbix_go_ibm_flash_v2/internal/parser"
	"zabbix_go_ibm_flash_v2/internal/ssh"
)

// DiscoveryItem representa un elemento descubierto por Zabbix
// Este formato es requerido por Zabbix para reglas de descubrimiento dinámico
type DiscoveryItem struct {
	// Macro que identifica el elemento descubierto
	Key string `json:"{#KEY}"`
	// Valor asociado a la macro
	Value string `json:"{#VALUE}"`
}

// DiscoveryItemExtended representa un elemento descubierto con múltiples propiedades
// Permite incluir información adicional sobre el elemento descubierto
type DiscoveryItemExtended struct {
	ID     string `json:"{#ID}"`
	Name   string `json:"{#NAME}"`
	Status string `json:"{#STATUS}"`
	Type   string `json:"{#TYPE}"`
}

// DiscoverDrives realiza el descubrimiento de unidades de disco en el sistema FlashSystem
// host: dirección IP o nombre de host del sistema FlashSystem
// user: nombre de usuario para autenticación
// pass: contraseña para autenticación
// Retorna un JSON con la información de las unidades descubiertas
func DiscoverDrives(host, user, pass string) (string, error) {
	// Establecer conexión SSH
	client, err := ssh.EstablishSSHConnection(host, user, pass)
	if err != nil {
		return "{}", fmt.Errorf("error estableciendo conexión SSH: %w", err)
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
		return "{}", fmt.Errorf("error ejecutando comando lsdrive: %w", err)
	}

	// Parsear la salida del comando
	records := parser.ParseDelimitedOutput(output, ':')

	// Verificar si hay unidades para procesar
	if len(records) == 0 {
		return `{"data":[]}`, nil
	}

	// Crear la estructura de datos para el descubrimiento
	var discoveryItems []DiscoveryItemExtended

	// Procesar cada registro de unidad
	for _, r := range records {
		// Crear un elemento de descubrimiento con múltiples propiedades
		item := DiscoveryItemExtended{
			ID:     escapeJSONValue(r["id"]),
			Name:   escapeJSONValue(r["name"]),
			Status: escapeJSONValue(r["status"]),
			Type:   escapeJSONValue(r["type"]),
		}

		// Agregar el elemento a la lista de descubrimiento
		discoveryItems = append(discoveryItems, item)
	}

	// Crear la estructura JSON de descubrimiento
	discoveryResult := struct {
		Data []DiscoveryItemExtended `json:"data"`
	}{
		Data: discoveryItems,
	}

	// Convertir la estructura a JSON
	jsonResult, err := json.Marshal(discoveryResult)
	if err != nil {
		return "{}", fmt.Errorf("error convirtiendo descubrimiento a JSON: %w", err)
	}

	// Retornar el JSON como string
	return string(jsonResult), nil
}

// DiscoverEnclosures realiza el descubrimiento de carcasas en el sistema FlashSystem
// host: dirección IP o nombre de host del sistema FlashSystem
// user: nombre de usuario para autenticación
// pass: contraseña para autenticación
// Retorna un JSON con la información de las carcasas descubiertas
func DiscoverEnclosures(host, user, pass string) (string, error) {
	// Establecer conexión SSH
	client, err := ssh.EstablishSSHConnection(host, user, pass)
	if err != nil {
		return "{}", fmt.Errorf("error estableciendo conexión SSH: %w", err)
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
		return "{}", fmt.Errorf("error ejecutando comando lsenclosure: %w", err)
	}

	// Parsear la salida del comando
	records := parser.ParseDelimitedOutput(output, ':')

	// Verificar si hay carcasas para procesar
	if len(records) == 0 {
		return `{"data":[]}`, nil
	}

	// Crear la estructura de datos para el descubrimiento
	var discoveryItems []DiscoveryItemExtended

	// Procesar cada registro de carcasa
	for _, r := range records {
		// Crear un elemento de descubrimiento con múltiples propiedades
		item := DiscoveryItemExtended{
			ID:     escapeJSONValue(r["id"]),
			Name:   escapeJSONValue(r["name"]),
			Status: escapeJSONValue(r["status"]),
			Type:   escapeJSONValue(r["type"]),
		}

		// Agregar el elemento a la lista de descubrimiento
		discoveryItems = append(discoveryItems, item)
	}

	// Crear la estructura JSON de descubrimiento
	discoveryResult := struct {
		Data []DiscoveryItemExtended `json:"data"`
	}{
		Data: discoveryItems,
	}

	// Convertir la estructura a JSON
	jsonResult, err := json.Marshal(discoveryResult)
	if err != nil {
		return "{}", fmt.Errorf("error convirtiendo descubrimiento a JSON: %w", err)
	}

	// Retornar el JSON como string
	return string(jsonResult), nil
}

// DiscoverPools realiza el descubrimiento de grupos de discos en el sistema FlashSystem
// host: dirección IP o nombre de host del sistema FlashSystem
// user: nombre de usuario para autenticación
// pass: contraseña para autenticación
// Retorna un JSON con la información de los grupos de discos descubiertos
func DiscoverPools(host, user, pass string) (string, error) {
	// Establecer conexión SSH
	client, err := ssh.EstablishSSHConnection(host, user, pass)
	if err != nil {
		return "{}", fmt.Errorf("error estableciendo conexión SSH: %w", err)
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
		return "{}", fmt.Errorf("error ejecutando comando lsmdiskgrp: %w", err)
	}

	// Parsear la salida del comando
	records := parser.ParseDelimitedOutput(output, ':')

	// Verificar si hay grupos de discos para procesar
	if len(records) == 0 {
		return `{"data":[]}`, nil
	}

	// Crear la estructura de datos para el descubrimiento
	var discoveryItems []DiscoveryItemExtended

	// Procesar cada registro de grupo de disco
	for _, r := range records {
		// Crear un elemento de descubrimiento con múltiples propiedades
		item := DiscoveryItemExtended{
			ID:     escapeJSONValue(r["id"]),
			Name:   escapeJSONValue(r["name"]),
			Status: escapeJSONValue(r["status"]),
			Type:   escapeJSONValue(r["type"]),
		}

		// Agregar el elemento a la lista de descubrimiento
		discoveryItems = append(discoveryItems, item)
	}

	// Crear la estructura JSON de descubrimiento
	discoveryResult := struct {
		Data []DiscoveryItemExtended `json:"data"`
	}{
		Data: discoveryItems,
	}

	// Convertir la estructura a JSON
	jsonResult, err := json.Marshal(discoveryResult)
	if err != nil {
		return "{}", fmt.Errorf("error convirtiendo descubrimiento a JSON: %w", err)
	}

	// Retornar el JSON como string
	return string(jsonResult), nil
}

// DiscoverVolumes realiza el descubrimiento de volúmenes virtuales en el sistema FlashSystem
// host: dirección IP o nombre de host del sistema FlashSystem
// user: nombre de usuario para autenticación
// pass: contraseña para autenticación
// Retorna un JSON con la información de los volúmenes descubiertos
func DiscoverVolumes(host, user, pass string) (string, error) {
	// Establecer conexión SSH
	client, err := ssh.EstablishSSHConnection(host, user, pass)
	if err != nil {
		return "{}", fmt.Errorf("error estableciendo conexión SSH: %w", err)
	}
	defer func() {
		if closeErr := client.Close(); closeErr != nil {
			log.Printf("Advertencia: Error al cerrar conexión SSH: %v", closeErr)
		}
	}()

	// Ejecutar comando para obtener información de volúmenes virtuales
	cmd := "svcinfo lsvdisk -delim :"
	output, err := ssh.ExecuteCommand(client, cmd)
	if err != nil {
		return "{}", fmt.Errorf("error ejecutando comando lsvdisk: %w", err)
	}

	// Parsear la salida del comando
	records := parser.ParseDelimitedOutput(output, ':')

	// Verificar si hay volúmenes para procesar
	if len(records) == 0 {
		return `{"data":[]}`, nil
	}

	// Crear la estructura de datos para el descubrimiento
	var discoveryItems []DiscoveryItemExtended

	// Procesar cada registro de volumen
	for _, r := range records {
		// Crear un elemento de descubrimiento con múltiples propiedades
		item := DiscoveryItemExtended{
			ID:     escapeJSONValue(r["id"]),
			Name:   escapeJSONValue(r["name"]),
			Status: escapeJSONValue(r["status"]),
			Type:   escapeJSONValue(r["vdisk_type"]),
		}

		// Agregar el elemento a la lista de descubrimiento
		discoveryItems = append(discoveryItems, item)
	}

	// Crear la estructura JSON de descubrimiento
	discoveryResult := struct {
		Data []DiscoveryItemExtended `json:"data"`
	}{
		Data: discoveryItems,
	}

	// Convertir la estructura a JSON
	jsonResult, err := json.Marshal(discoveryResult)
	if err != nil {
		return "{}", fmt.Errorf("error convirtiendo descubrimiento a JSON: %w", err)
	}

	// Retornar el JSON como string
	return string(jsonResult), nil
}

// DiscoverNodes realiza el descubrimiento de nodos en el sistema FlashSystem
// host: dirección IP o nombre de host del sistema FlashSystem
// user: nombre de usuario para autenticación
// pass: contraseña para autenticación
// Retorna un JSON con la información de los nodos descubiertos
func DiscoverNodes(host, user, pass string) (string, error) {
	// Establecer conexión SSH
	client, err := ssh.EstablishSSHConnection(host, user, pass)
	if err != nil {
		return "{}", fmt.Errorf("error estableciendo conexión SSH: %w", err)
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
		return "{}", fmt.Errorf("error ejecutando comando lsnode: %w", err)
	}

	// Parsear la salida del comando
	records := parser.ParseDelimitedOutput(output, ':')

	// Verificar si hay nodos para procesar
	if len(records) == 0 {
		return `{"data":[]}`, nil
	}

	// Crear la estructura de datos para el descubrimiento
	var discoveryItems []DiscoveryItemExtended

	// Procesar cada registro de nodo
	for _, r := range records {
		// Crear un elemento de descubrimiento con múltiples propiedades
		item := DiscoveryItemExtended{
			ID:     escapeJSONValue(r["id"]),
			Name:   escapeJSONValue(r["name"]),
			Status: escapeJSONValue(r["status"]),
			Type:   escapeJSONValue(r["type"]),
		}

		// Agregar el elemento a la lista de descubrimiento
		discoveryItems = append(discoveryItems, item)
	}

	// Crear la estructura JSON de descubrimiento
	discoveryResult := struct {
		Data []DiscoveryItemExtended `json:"data"`
	}{
		Data: discoveryItems,
	}

	// Convertir la estructura a JSON
	jsonResult, err := json.Marshal(discoveryResult)
	if err != nil {
		return "{}", fmt.Errorf("error convirtiendo descubrimiento a JSON: %w", err)
	}

	// Retornar el JSON como string
	return string(jsonResult), nil
}

// DiscoverFCPorts realiza el descubrimiento de puertos Fibre Channel en el sistema FlashSystem
// host: dirección IP o nombre de host del sistema FlashSystem
// user: nombre de usuario para autenticación
// pass: contraseña para autenticación
// Retorna un JSON con la información de los puertos FC descubiertos
func DiscoverFCPorts(host, user, pass string) (string, error) {
	// Establecer conexión SSH
	client, err := ssh.EstablishSSHConnection(host, user, pass)
	if err != nil {
		return "{}", fmt.Errorf("error estableciendo conexión SSH: %w", err)
	}
	defer func() {
		if closeErr := client.Close(); closeErr != nil {
			log.Printf("Advertencia: Error al cerrar conexión SSH: %v", closeErr)
		}
	}()

	// Ejecutar comando para obtener información de puertos FC
	cmd := "svcinfo lsportfc -delim :"
	output, err := ssh.ExecuteCommand(client, cmd)
	if err != nil {
		return "{}", fmt.Errorf("error ejecutando comando lsportfc: %w", err)
	}

	// Parsear la salida del comando
	records := parser.ParseDelimitedOutput(output, ':')

	// Verificar si hay puertos FC para procesar
	if len(records) == 0 {
		return `{"data":[]}`, nil
	}

	// Crear la estructura de datos para el descubrimiento
	var discoveryItems []DiscoveryItemExtended

	// Procesar cada registro de puerto FC
	for _, r := range records {
		// Crear un elemento de descubrimiento con múltiples propiedades
		item := DiscoveryItemExtended{
			ID:     escapeJSONValue(r["id"]),
			Name:   escapeJSONValue(r["WWPN"]),
			Status: escapeJSONValue(r["status"]),
			Type:   escapeJSONValue(r["type"]),
		}

		// Agregar el elemento a la lista de descubrimiento
		discoveryItems = append(discoveryItems, item)
	}

	// Crear la estructura JSON de descubrimiento
	discoveryResult := struct {
		Data []DiscoveryItemExtended `json:"data"`
	}{
		Data: discoveryItems,
	}

	// Convertir la estructura a JSON
	jsonResult, err := json.Marshal(discoveryResult)
	if err != nil {
		return "{}", fmt.Errorf("error convirtiendo descubrimiento a JSON: %w", err)
	}

	// Retornar el JSON como string
	return string(jsonResult), nil
}

// escapeJSONValue escapa caracteres especiales en un valor para que sea seguro en JSON
// value: el valor a escapar
// Retorna el valor con caracteres especiales escapados
func escapeJSONValue(value string) string {
	// Reemplazar caracteres que podrían causar problemas en JSON
	value = strings.ReplaceAll(value, `"`, `\"`)  // Comillas dobles
	value = strings.ReplaceAll(value, `\`, `\\`)  // Backslash
	value = strings.ReplaceAll(value, "\n", `\n`) // Nueva línea
	value = strings.ReplaceAll(value, "\r", `\r`) // Retorno de carro
	value = strings.ReplaceAll(value, "\t", `\t`) // Tabulador

	return value
}

// DiscoverFlashCopyRelationships realiza el descubrimiento de relaciones FlashCopy
// host: dirección IP o nombre de host del sistema FlashSystem
// user: nombre de usuario para autenticación
// pass: contraseña para autenticación
// Retorna un JSON con la información de las relaciones FlashCopy descubiertas
func DiscoverFlashCopyRelationships(host, user, pass string) (string, error) {
	// Establecer conexión SSH
	client, err := ssh.EstablishSSHConnection(host, user, pass)
	if err != nil {
		return "{}", fmt.Errorf("error estableciendo conexión SSH: %w", err)
	}
	defer func() {
		if closeErr := client.Close(); closeErr != nil {
			log.Printf("Advertencia: Error al cerrar conexión SSH: %v", closeErr)
		}
	}()

	// Ejecutar comando para obtener información de relaciones FlashCopy
	cmd := "svcinfo lsfcmap -delim :"
	output, err := ssh.ExecuteCommand(client, cmd)
	if err != nil {
		return "{}", fmt.Errorf("error ejecutando comando lsfcmap: %w", err)
	}

	// Parsear la salida del comando
	records := parser.ParseDelimitedOutput(output, ':')

	// Verificar si hay relaciones FlashCopy para procesar
	if len(records) == 0 {
		return `{"data":[]}`, nil
	}

	// Crear la estructura de datos para el descubrimiento
	var discoveryItems []DiscoveryItemExtended

	// Procesar cada registro de relación FlashCopy
	for _, r := range records {
		// Crear un elemento de descubrimiento con múltiples propiedades
		item := DiscoveryItemExtended{
			ID:     escapeJSONValue(r["id"]),
			Name:   escapeJSONValue(r["source_vdisk_name"] + "->" + r["target_vdisk_name"]),
			Status: escapeJSONValue(r["status"]),
			Type:   escapeJSONValue(r["copy_rate"]),
		}

		// Agregar el elemento a la lista de descubrimiento
		discoveryItems = append(discoveryItems, item)
	}

	// Crear la estructura JSON de descubrimiento
	discoveryResult := struct {
		Data []DiscoveryItemExtended `json:"data"`
	}{
		Data: discoveryItems,
	}

	// Convertir la estructura a JSON
	jsonResult, err := json.Marshal(discoveryResult)
	if err != nil {
		return "{}", fmt.Errorf("error convirtiendo descubrimiento a JSON: %w", err)
	}

	// Retornar el JSON como string
	return string(jsonResult), nil
}

// DiscoverReplicationRelationships realiza el descubrimiento de relaciones de replicación
// host: dirección IP o nombre de host del sistema FlashSystem
// user: nombre de usuario para autenticación
// pass: contraseña para autenticación
// Retorna un JSON con la información de las relaciones de replicación descubiertas
func DiscoverReplicationRelationships(host, user, pass string) (string, error) {
	// Establecer conexión SSH
	client, err := ssh.EstablishSSHConnection(host, user, pass)
	if err != nil {
		return "{}", fmt.Errorf("error estableciendo conexión SSH: %w", err)
	}
	defer func() {
		if closeErr := client.Close(); closeErr != nil {
			log.Printf("Advertencia: Error al cerrar conexión SSH: %v", closeErr)
		}
	}()

	// Ejecutar comando para obtener información de relaciones de replicación
	cmd := "svcinfo lsrcrelationship -delim :"
	output, err := ssh.ExecuteCommand(client, cmd)
	if err != nil {
		return "{}", fmt.Errorf("error ejecutando comando lsrcrelationship: %w", err)
	}

	// Parsear la salida del comando
	records := parser.ParseDelimitedOutput(output, ':')

	// Verificar si hay relaciones de replicación para procesar
	if len(records) == 0 {
		return `{"data":[]}`, nil
	}

	// Crear la estructura de datos para el descubrimiento
	var discoveryItems []DiscoveryItemExtended

	// Procesar cada registro de relación de replicación
	for _, r := range records {
		// Crear un elemento de descubrimiento con múltiples propiedades
		item := DiscoveryItemExtended{
			ID:     escapeJSONValue(r["id"]),
			Name:   escapeJSONValue(r["master_cluster_name"] + "->" + r["aux_cluster_name"]),
			Status: escapeJSONValue(r["state"]),
			Type:   escapeJSONValue(r["copy_type"]),
		}

		// Agregar el elemento a la lista de descubrimiento
		discoveryItems = append(discoveryItems, item)
	}

	// Crear la estructura JSON de descubrimiento
	discoveryResult := struct {
		Data []DiscoveryItemExtended `json:"data"`
	}{
		Data: discoveryItems,
	}

	// Convertir la estructura a JSON
	jsonResult, err := json.Marshal(discoveryResult)
	if err != nil {
		return "{}", fmt.Errorf("error convirtiendo descubrimiento a JSON: %w", err)
	}

	// Retornar el JSON como string
	return string(jsonResult), nil
}
