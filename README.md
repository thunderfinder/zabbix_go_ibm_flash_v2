# zabbix_go_ibm_flash_v2

```markdown
# IBM FlashSystem Zabbix Collector

Colector de métricas para sistemas IBM FlashSystem e IBM Spectrum Virtualize que se integra con Zabbix para monitoreo de infraestructura de almacenamiento.

## Descripción del proyecto

Este proyecto proporciona un colector escrito en Go que se conecta a sistemas de almacenamiento IBM FlashSystem e IBM Spectrum Virtualize mediante SSH para recopilar métricas de estado, rendimiento y disponibilidad. El colector ejecuta comandos CLI del sistema de almacenamiento (`svcinfo`) y proporciona los resultados en un formato compatible con Zabbix, permitiendo el monitoreo automatizado de componentes críticos como discos, carcasas, grupos de discos, volúmenes, nodos, puertos FC y relaciones de replicación.

## Características principales

* **Monitoreo de componentes de IBM FlashSystem**: Discos, carcasas, grupos de discos, volúmenes, nodos, puertos FC
* **Descubrimiento dinámico (LLD)**: Detección automática de componentes para creación de reglas de monitoreo en Zabbix
* **Métricas de rendimiento**: IOPS, latencia, capacidad y uso de almacenamiento
* **Integración con Zabbix**: Compatible con External Scripts y UserParameters
* **Seguridad SSH**: Conexión segura con autenticación por contraseña o llaves SSH
* **Arquitectura modular**: Código Go bien estructurado y documentado
* **Soporte para múltiples versiones**: Compatible con IBM Spectrum Virtualize v8.7

## Arquitectura general

El sistema consta de:

1. **Binario Go**: Programa principal que se comunica con el sistema de almacenamiento
2. **Comandos CLI del storage**: Utiliza comandos `svcinfo` para obtener información del sistema
3. **Integración con Zabbix**: El script actúa como External Script o UserParameter en Zabbix
4. **Formato de salida**: Datos estructurados en JSON para discovery y valores simples para métricas

## Requisitos del sistema

* **Go**: Versión 1.22 o superior para compilar el proyecto
* **Zabbix**: Versión 7.2 o superior (funciona con Zabbix Server o Proxy)
* **Acceso SSH**: Conexión SSH al sistema IBM FlashSystem con un usuario autorizado
* **Usuario de monitoreo**: Cuenta en el sistema de almacenamiento con permisos de solo lectura
* **Puerto SSH**: Puerto 22 accesible desde el servidor Zabbix

## Instalación paso a paso

### 1. Clonar el repositorio

```bash
git clone https://github.com/tu-usuario/flashsystem-zabbix-collector.git
cd flashsystem-zabbix-collector
```

### 2. Compilar el binario Go

```bash
go mod tidy
go build -o flashsystem_collector .
```

### 3. Copiar archivos al servidor Zabbix

```bash
# Copiar el binario al servidor Zabbix
scp flashsystem_collector zabbix-server:/usr/local/bin/

# Copiar el script de wrapper
scp scripts/flashsystem_collector.sh zabbix-server:/usr/lib/zabbix/externalscripts/
```

### 4. Configurar permisos

```bash
# En el servidor Zabbix
sudo chown root:zabbix /usr/local/bin/flashsystem_collector
sudo chmod 750 /usr/local/bin/flashsystem_collector

sudo chown zabbix:zabbix /usr/lib/zabbix/externalscripts/flashsystem_collector.sh
sudo chmod 755 /usr/lib/zabbix/externalscripts/flashsystem_collector.sh
```

## Configuración

### 1. Configurar acceso SSH al sistema de almacenamiento

Crear un usuario de monitoreo en el sistema IBM FlashSystem con permisos de solo lectura.

### 2. Configurar credenciales (opcional)

Para mayor seguridad, puedes crear un archivo de credenciales:

```bash
# Crear directorio de configuración
sudo mkdir -p /etc/zabbix/flashsystem
sudo chown zabbix:zabbix /etc/zabbix/flashsystem
sudo chmod 700 /etc/zabbix/flashsystem

# Crear archivo de credenciales
sudo tee /etc/zabbix/flashsystem/credentials << EOF
# Credenciales para IBM FlashSystem
# Formato: host:usuario:contraseña
10.10.10.50:monitor:tu_contraseña_segura
EOF

sudo chmod 600 /etc/zabbix/flashsystem/credentials
```

### 3. Configurar Zabbix

Editar el archivo de configuración de Zabbix Server o Proxy para asegurar que los External Scripts estén habilitados:

```
ExternalScripts=/usr/lib/zabbix/externalscripts
```

## Uso del script

El script se puede ejecutar manualmente para probar su funcionamiento:

```bash
# Ejemplo: Obtener estado de un disco específico
/usr/lib/zabbix/externalscripts/flashsystem_collector.sh 10.10.10.50 monitor contraseña_secreta getdrivestatus 1 12

# Ejemplo: Descubrir todos los grupos de discos
/usr/lib/zabbix/externalscripts/flashsystem_collector.sh 10.10.10.50 monitor contraseña_secreta discoverpools

# Ejemplo: Obtener uso de un grupo de discos específico
/usr/lib/zabbix/externalscripts/flashsystem_collector.sh 10.10.10.50 monitor contraseña_secreta getpoolusage Pool0
```

### Parámetros del script

* `$1`: Dirección IP o hostname del sistema IBM FlashSystem
* `$2`: Usuario para autenticación SSH
* `$3`: Contraseña para autenticación SSH
* `$4`: Comando a ejecutar (ver sección de comandos disponibles)
* `$5+`: Parámetros adicionales según el comando

### Comandos disponibles

* `discoverdrives` - Descubre unidades de disco
* `discoverenclosures` - Descubre carcasas
* `discoverpools` - Descubre grupos de discos
* `discovervolumes` - Descubre volúmenes virtuales
* `discovernodes` - Descubre nodos
* `discoverfcports` - Descubre puertos FC
* `discoverflashcopy` - Descubre relaciones FlashCopy
* `discoverreplication` - Descubre relaciones de replicación
* `getdrivestatus` - Obtiene estado de unidad
* `getenclosurestatus` - Obtiene estado de carcasa
* `getbatterystatus` - Obtiene estado de batería
* `getpoolusage` - Obtiene uso de grupo de discos
* `getiops` - Obtiene IOPS del sistema
* `getvolumestatus` - Obtiene estado de volumen
* `getnodestatus` - Obtiene estado de nodo
* `getclusterstatus` - Obtiene estado del cluster

## Integración con Zabbix

### 1. Crear UserParameter (opcional)

Agregar al archivo de configuración de Zabbix Agent o Proxy:

```
# Discovery de unidades
UserParameter=flashsystem.discovery.drives[*],/usr/lib/zabbix/externalscripts/flashsystem_collector.sh $1 $2 $3 discoverdrives

# Métricas de estado
UserParameter=flashsystem.drive.status[*],/usr/lib/zabbix/externalscripts/flashsystem_collector.sh $1 $2 $3 getdrivestatus $4 $5
UserParameter=flashsystem.pool.usage[*],/usr/lib/zabbix/externalscripts/flashsystem_collector.sh $1 $2 $3 getpoolusage $4
UserParameter=flashsystem.system.iops[*],/usr/lib/zabbix/externalscripts/flashsystem_collector.sh $1 $2 $3 getiops
```

### 2. Crear Items y Reglas de Discovery en Zabbix

#### Item de ejemplo para IOPS del sistema:
* **Tipo**: External check
* **Key**: `flashsystem_collector.sh[{HOST.IP},monitor,contraseña,getiops]`
* **Tipo de información**: Numérico (entero sin signo)
* **Intervalo de actualización**: 60s

#### Regla de Discovery para grupos de discos:
* **Tipo**: External check
* **Key**: `flashsystem_collector.sh[{HOST.IP},monitor,contraseña,discoverpools]`
* **Tipo de prototipo de item**: External check
* **Key de prototipo**: `flashsystem_collector.sh[{HOST.IP},monitor,contraseña,getpoolusage,{#NAME}]`

## Ejemplos de monitoreo

### Métricas disponibles

* **Estado del sistema**: Cluster health, nodos operativos
* **Grupos de discos**: Capacidad total, espacio libre, uso porcentual
* **Discos físicos**: Estado online/offline, temperatura, errores
* **Volúmenes virtuales**: Estado, tamaño, tipo
* **Carcasas**: Estado físico, temperatura, alimentación
* **Puertos FC**: Estado, WWPN, velocidad de conexión
* **Relaciones FlashCopy**: Estado de sincronización
* **Replicación remota**: Estado de la relación, latencia
* **Rendimiento**: IOPS totales, latencia promedio, throughput

### Discovery de componentes

* **Discos**: `{#ID}`, `{#ENCLOSUREID}`, `{#SLOTID}`, `{#STATUS}`
* **Grupos de discos**: `{#ID}`, `{#NAME}`, `{#STATUS}`
* **Volúmenes**: `{#ID}`, `{#NAME}`, `{#STATUS}`, `{#VDISK_TYPE}`
* **Nodos**: `{#ID}`, `{#NAME}`, `{#STATUS}`, `{#TYPE}`
* **Puertos FC**: `{#ID}`, `{#WWPN}`, `{#STATUS}`, `{#TYPE}`

## Estructura del proyecto

```
zabbix_go_ibm_flash_v2/
├── go.mod                 # Dependencias del proyecto Go
├── main.go                # Punto de entrada principal del programa
├── internal/
│   ├── ssh/              # Gestión de conexiones SSH
│   │   └── ssh.go
│   ├── parser/           # Análisis de salida de comandos SVC
│   │   └── parser.go
│   ├── monitor/          # Funciones de monitoreo específicas
│   │   └── monitor.go
│   ├── discovery/        # Funciones de descubrimiento LLD
│   │   └── discovery.go
│   └── utils/            # Funciones de utilidad general
│       └── utils.go
├── scripts/
│   └── flashsystem_collector.sh  # Script wrapper para Zabbix
└── README.md             # Documentación del proyecto
```

## Seguridad

### Buenas prácticas

* **Usar cuenta de solo lectura**: El usuario SSH debe tener permisos de solo lectura
* **Restringir acceso SSH**: Limitar el acceso SSH solo a direcciones IP confiables
* **Proteger credenciales**: No almacenar contraseñas en scripts visibles
* **Auditar conexiones**: Habilitar logging de conexiones SSH
* **Actualizar regularmente**: Mantener el software actualizado

### Consideraciones de seguridad

* El script ejecuta comandos de solo lectura (`svcinfo`)
* No modifica el estado del sistema de almacenamiento
* Se recomienda usar autenticación por llaves SSH en lugar de contraseñas
* El script no almacena credenciales permanentemente

## Troubleshooting

### Problemas comunes

**Error de conexión SSH**
* Verificar que el puerto 22 esté accesible
* Confirmar que el usuario y contraseña sean correctos
* Asegurarse de que el firewall no bloquee la conexión

**Error de permisos**
* Verificar que el usuario Zabbix pueda ejecutar el script
* Confirmar que el binario tenga permisos de ejecución

**Comando no reconocido**
* Verificar que el comando esté en la lista de comandos permitidos
* Revisar la sintaxis del comando

**Timeout en ejecución**
* Aumentar el timeout en el script si el sistema es lento
* Verificar la latencia de red entre servidores

### Logging

El script registra eventos en `/var/log/zabbix/flashsystem_collector.log` para facilitar la depuración.

## Contribuciones

Las contribuciones son bienvenidas. Para contribuir:

1. Haz un fork del proyecto
2. Crea una rama para tu característica (`git checkout -b feature/nueva-caracteristica`)
3. Realiza tus cambios
4. Asegúrate de que el código cumple con las convenciones
5. Escribe pruebas si es necesario
6. Haz commit de tus cambios (`git commit -am 'Agrega nueva característica'`)
7. Sube tu rama (`git push origin feature/nueva-caracteristica`)
8. Abre un Pull Request

## Licencia


```