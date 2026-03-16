// Define el módulo Go para el proyecto de monitoreo de IBM FlashSystem
// Este archivo gestiona las dependencias externas necesarias para la aplicación
module flashsystem_zabbix

// Especifica la versión mínima de Go requerida para compilar este proyecto
go 1.22

// Lista de dependencias externas utilizadas por el proyecto
require (
	// Biblioteca criptográfica de Go para operaciones de seguridad SSH
	golang.org/x/crypto v0.22.0
)