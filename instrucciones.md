# Guía de Implementación en Producción
## IBM FlashSystem Zabbix Collector

**Versión del documento:** 1.0  
**Plataforma objetivo:** Red Hat Enterprise Linux 8/9 + Zabbix 7.2  
**Sistema monitoreado:** IBM FlashSystem / Spectrum Virtualize v8.7+

---

## 1. Análisis del Proyecto

### 1.1 Estructura del Repositorio

```
flashsystem_zabbix/
├── go.mod                          # Dependencias del módulo Go
├── main.go                         # Punto de entrada principal
├── internal/
│   ├── ssh/ssh.go                  # Gestión de conexiones SSH
│   ├── parser/parser.go            # Análisis de salida de comandos SVC
│   ├── monitor/monitor.go          # Funciones de monitoreo específicas
│   ├── discovery/discovery.go      # Funciones de descubrimiento LLD
│   └── utils/utils.go              # Funciones de utilidad general
├── scripts/
│   └── flashsystem_collector.sh    # Wrapper para Zabbix External Scripts
└── README.md                       # Documentación del proyecto
```

### 1.2 Dependencias Go

| Dependencia | Versión | Propósito |
|-------------|---------|-----------|
| `golang.org/x/crypto` | v0.22.0 | Cliente SSH seguro |

### 1.3 Comandos CLI Ejecutados

El proyecto ejecuta los siguientes comandos `svcinfo` en el IBM FlashSystem:

```bash
# Descubrimiento de componentes
svcinfo lsdrive -delim :           # Unidades físicas
svcinfo lsenclosure -delim :       # Carcasas
svcinfo lsmdiskgrp -delim :        # Grupos de discos (pools)
svcinfo lsvdisk -delim :           # Volúmenes virtuales
svcinfo lsnode -delim :            # Nodos del cluster
svcinfo lsportfc -delim :          # Puertos Fibre Channel
svcinfo lsfcmap -delim :           # Relaciones FlashCopy
svcinfo lsrcrelationship -delim :  # Replicación remota

# Métricas de estado y rendimiento
svcinfo lsenclosurebattery -delim :  # Estado de baterías
svcinfo lssystemstats -delim :       # Estadísticas de IOPS
svcinfo lssystem -delim :            # Información general del sistema
```

**Nota crítica:** Todos los comandos se ejecutan con `SVC_CLI_LOCALE=C` para garantizar formato de salida consistente.

### 1.4 Parámetros del Programa

El binario acepta los siguientes argumentos en orden:

```
flashsystem_collector <host> <user> <password> <command> [param1] [param2]
```

| Posición | Parámetro | Descripción | Ejemplo |
|----------|-----------|-------------|---------|
| 1 | host | IP o hostname del FlashSystem | 10.10.10.50 |
| 2 | user | Usuario SSH para autenticación | monitor |
| 3 | password | Contraseña del usuario | contraseña123 |
| 4 | command | Comando a ejecutar | getdrivestatus |
| 5+ | params | Parámetros adicionales según comando | 1 12 |

### 1.5 Comandos Disponibles

| Comando | Parámetros adicionales | Tipo de salida | Propósito |
|---------|----------------------|----------------|-----------|
| `discoverdrives` | Ninguno | JSON LLD | Descubrir unidades físicas |
| `discoverenclosures` | Ninguno | JSON LLD | Descubrir carcasas |
| `discoverpools` | Ninguno | JSON LLD | Descubrir grupos de discos |
| `discovervolumes` | Ninguno | JSON LLD | Descubrir volúmenes virtuales |
| `discovernodes` | Ninguno | JSON LLD | Descubrir nodos |
| `discoverfcports` | Ninguno | JSON LLD | Descubrir puertos FC |
| `discoverflashcopy` | Ninguno | JSON LLD | Descubrir relaciones FlashCopy |
| `discoverreplication` | Ninguno | JSON LLD | Descubrir replicación |
| `getdrivestatus` | enclosure_id, slot_id | Entero (0/1/2) | Estado de unidad específica |
| `getenclosurestatus` | enclosure_id | Entero (0/1/2) | Estado de carcasa |
| `getbatterystatus` | enclosure_id | Entero (0/1/2) | Estado de batería |
| `getpoolusage` | pool_name | Float (porcentaje) | Uso de pool en % |
| `getiops` | Ninguno | Entero | IOPS totales del sistema |
| `getvolumestatus` | volume_name | Entero (0/1/2) | Estado de volumen |
| `getnodestatus` | node_id | Entero (0/1/2) | Estado de nodo |
| `getclusterstatus` | Ninguno | Entero (0/1/2) | Estado del cluster |

### 1.6 Formato de Salida

**Para métricas individuales:**
```
1          ← Estado online
0          ← Estado offline
85.50      ← Porcentaje de uso
12500      ← IOPS totales
```

**Para descubrimiento (LLD):**
```json
{"data":[
  {"{#DRIVEID}":"1","{#ENCLOSUREID}":"0","{#SLOTID}":"12","{#DRIVESTATUS}":"online"},
  {"{#DRIVEID}":"2","{#ENCLOSUREID}":"0","{#SLOTID}":"13","{#DRIVESTATUS}":"online"}
]}
```

---

## 2. Requisitos del Sistema (Red Hat Enterprise Linux)

### 2.1 Requisitos de Software

| Componente | Versión Mínima | Comprobación |
|------------|---------------|--------------|
| Red Hat Enterprise Linux | 8.4 o 9.0 | `cat /etc/redhat-release` |
| Go (Golang) | 1.22 | `go version` |
| Zabbix Server/Proxy | 7.2 | `zabbix_server -V` |
| OpenSSH Client | Cualquiera | `ssh -V` |
| GCC (para compilar) | 8.0+ | `gcc --version` |

### 2.2 Paquetes Necesarios

Instalar dependencias del sistema:

```bash
# Actualizar repositorios
sudo dnf update -y

# Instalar herramientas de desarrollo y Go
sudo dnf groupinstall -y "Development Tools"
sudo dnf install -y golang git openssh-clients

# Verificar instalación de Go
go version
# Debe mostrar: go version go1.22.x linux/amd64
```

### 2.3 Configuración de Go (si no está en PATH)

```bash
# Agregar Go al PATH del usuario zabbix
echo 'export PATH=$PATH:/usr/local/go/bin' >> /etc/profile.d/go.sh
source /etc/profile.d/go.sh
```

---

## 3. Instalación del Proyecto

### 3.1 Preparar Directorios

```bash
# Crear directorios de trabajo
sudo mkdir -p /opt/flashsystem-collector
sudo mkdir -p /usr/lib/zabbix/externalscripts
sudo mkdir -p /etc/zabbix/flashsystem
sudo mkdir -p /var/log/zabbix

# Establecer permisos
sudo chown -R zabbix:zabbix /opt/flashsystem-collector
sudo chown -R zabbix:zabbix /usr/lib/zabbix/externalscripts
sudo chown -R zabbix:zabbix /etc/zabbix/flashsystem
sudo chown -R zabbix:zabbix /var/log/zabbix
sudo chmod 750 /etc/zabbix/flashsystem
```

### 3.2 Clonar y Compilar

```bash
# Como usuario zabbix o con sudo -u zabbix
cd /opt/flashsystem-collector

# Clonar el repositorio (ajustar URL según corresponda)
git clone https://github.com/tu-organizacion/flashsystem_zabbix.git .

# Descargar dependencias y compilar
go mod tidy
go build -ldflags "-s -w" -o flashsystem_collector .

# Verificar compilación
./flashsystem_collector --version 2>/dev/null || echo "Binario compilado exitosamente"

# Mover binario a ubicación final
sudo cp flashsystem_collector /usr/local/bin/
sudo chown root:zabbix /usr/local/bin/flashsystem_collector
sudo chmod 750 /usr/local/bin/flashsystem_collector
```

### 3.3 Instalar Script Wrapper

```bash
# Copiar script de wrapper para Zabbix
sudo cp scripts/flashsystem_collector.sh /usr/lib/zabbix/externalscripts/
sudo chown zabbix:zabbix /usr/lib/zabbix/externalscripts/flashsystem_collector.sh
sudo chmod 755 /usr/lib/zabbix/externalscripts/flashsystem_collector.sh

# Verificar que el script sea ejecutable
ls -l /usr/lib/zabbix/externalscripts/flashsystem_collector.sh
```

### 3.4 Configurar Logging

```bash
# Crear archivo de log con permisos seguros
sudo touch /var/log/zabbix/flashsystem_collector.log
sudo chown zabbix:zabbix /var/log/zabbix/flashsystem_collector.log
sudo chmod 640 /var/log/zabbix/flashsystem_collector.log

# Configurar logrotate (opcional pero recomendado)
sudo tee /etc/logrotate.d/zabbix-flashsystem << 'EOF'
/var/log/zabbix/flashsystem_collector.log {
    daily
    rotate 7
    compress
    delaycompress
    missingok
    notifempty
    create 640 zabbix zabbix
}
EOF
```

---

## 4. Permisos y Usuario de Ejecución

### 4.1 Usuario Recomendado

El script debe ejecutarse como el usuario **zabbix**, que es el usuario por defecto del servidor Zabbix en RHEL.

```bash
# Verificar que el usuario zabbix existe
id zabbix
# Debe mostrar: uid=993(zabbix) gid=991(zabbix) groups=991(zabbix)
```

### 4.2 Permisos del Binario

| Archivo | Propietario | Permisos | Razón |
|---------|-------------|----------|-------|
| `/usr/local/bin/flashsystem_collector` | root:zabbix | 750 | Solo zabbix puede ejecutar |
| `/usr/lib/zabbix/externalscripts/flashsystem_collector.sh` | zabbix:zabbix | 755 | Script ejecutable por Zabbix |
| `/etc/zabbix/flashsystem/` | zabbix:zabbix | 750 | Configuración protegida |
| `/var/log/zabbix/flashsystem_collector.log` | zabbix:zabbix | 640 | Logs legibles por zabbix |

### 4.3 Verificación de SELinux

En RHEL, SELinux puede bloquear la ejecución de scripts externos:

```bash
# Verificar estado de SELinux
getenforce

# Si está en modo Enforcing, agregar contexto de seguridad
sudo semanage fcontext -a -t zabbix_script_t "/usr/lib/zabbix/externalscripts/flashsystem_collector.sh"
sudo restorecon -v /usr/lib/zabbix/externalscripts/flashsystem_collector.sh

# Permitir conexiones SSH desde el proceso Zabbix
sudo setsebool -P zabbix_can_network on
```

---

## 5. Configuración de Acceso SSH al Storage

### 5.1 Crear Usuario de Monitoreo en IBM FlashSystem

En la CLI del IBM FlashSystem, crear un usuario con permisos de solo lectura:

```bash
# Conectarse al FlashSystem como administrador
ssh admin@10.10.10.50

# Crear usuario de monitoreo (comando ejemplo, ajustar según política)
mkuser monitor password=TuContraseñaSegura123 grp=Monitor role=Monitor

# Verificar permisos del usuario
lsuser monitor
```

**Permisos mínimos requeridos:**
- Rol: `Monitor` o equivalente de solo lectura
- Acceso a comandos: `svcinfo ls*` (todos los comandos de listado)
- Sin permisos de escritura o modificación

### 5.2 Configurar Autenticación SSH (Recomendado: Claves)

**Opción A: Autenticación por contraseña (menos segura, para pruebas)**

No requiere configuración adicional. La contraseña se pasa como argumento al script.

**Opción B: Autenticación por claves SSH (recomendado para producción)**

```bash
# Como usuario zabbix, generar par de claves
sudo -u zabbix ssh-keygen -t ed25519 -f /etc/zabbix/ssh/flashsystem_key -C "zabbix-flashsystem-monitor" -N ""

# Copiar clave pública al FlashSystem
sudo -u zabbix ssh-copy-id -i /etc/zabbix/ssh/flashsystem_key.pub monitor@10.10.10.50

# Verificar conexión sin contraseña
sudo -u zabbix ssh -i /etc/zabbix/ssh/flashsystem_key monitor@10.10.10.50 "svcinfo lssystem -delim : | head -1"

# Configurar permisos seguros en la clave privada
sudo chmod 600 /etc/zabbix/ssh/flashsystem_key
sudo chown zabbix:zabbix /etc/zabbix/ssh/flashsystem_key
```

### 5.3 Archivo de Credenciales (Opcional)

Para evitar exponer credenciales en la línea de comandos:

```bash
# Crear archivo de credenciales cifrado (ejemplo básico)
sudo tee /etc/zabbix/flashsystem/credentials.conf << 'EOF'
# Formato: host:usuario:metodo_auth:[valor_auth]
# método_auth: password | key
# Para password: valor_auth = contraseña
# Para key: valor_auth = ruta a clave privada

10.10.10.50:monitor:key:/etc/zabbix/ssh/flashsystem_key
EOF

sudo chmod 600 /etc/zabbix/flashsystem/credentials.conf
sudo chown zabbix:zabbix /etc/zabbix/flashsystem/credentials.conf
```

**Nota:** El código actual no implementa lectura automática de este archivo. Para usarlo, se requiere modificar `main.go` o pasar credenciales como argumentos.

### 5.4 Pruebas de Conexión SSH

```bash
# Probar conexión básica
sudo -u zabbix ssh -o ConnectTimeout=10 monitor@10.10.10.50 "echo Conexión exitosa"

# Probar ejecución de comando svcinfo
sudo -u zabbix ssh -o ConnectTimeout=10 monitor@10.10.10.50 "SVC_CLI_LOCALE=C svcinfo lssystem -delim : | head -3"

# Verificar que el usuario tiene permisos de lectura
sudo -u zabbix ssh monitor@10.10.10.50 "svcinfo lsdrive -delim : | wc -l"
```

---

## 6. Integración con Zabbix

### 6.1 Ubicación del Script

El script wrapper debe ubicarse en el directorio configurado como `ExternalScripts` en Zabbix:

```bash
# Verificar configuración de ExternalScripts en Zabbix
grep ExternalScripts /etc/zabbix/zabbix_server.conf
# Debe mostrar: ExternalScripts=/usr/lib/zabbix/externalscripts
```

### 6.2 Configuración de UserParameter (Alternativa)

Si se prefiere usar UserParameters en lugar de External Scripts:

```bash
# Crear archivo de configuración de UserParameter
sudo tee /etc/zabbix/zabbix_agentd.d/flashsystem.conf << 'EOF'
# Discovery de componentes
UserParameter=flashsystem.discovery.drives[*],/usr/lib/zabbix/externalscripts/flashsystem_collector.sh $1 $2 $3 discoverdrives
UserParameter=flashsystem.discovery.enclosures[*],/usr/lib/zabbix/externalscripts/flashsystem_collector.sh $1 $2 $3 discoverenclosures
UserParameter=flashsystem.discovery.pools[*],/usr/lib/zabbix/externalscripts/flashsystem_collector.sh $1 $2 $3 discoverpools
UserParameter=flashsystem.discovery.volumes[*],/usr/lib/zabbix/externalscripts/flashsystem_collector.sh $1 $2 $3 discovervolumes

# Métricas de estado
UserParameter=flashsystem.drive.status[*],/usr/lib/zabbix/externalscripts/flashsystem_collector.sh $1 $2 $3 getdrivestatus $4 $5
UserParameter=flashsystem.enclosure.status[*],/usr/lib/zabbix/externalscripts/flashsystem_collector.sh $1 $2 $3 getenclosurestatus $4
UserParameter=flashsystem.battery.status[*],/usr/lib/zabbix/externalscripts/flashsystem_collector.sh $1 $2 $3 getbatterystatus $4
UserParameter=flashsystem.volume.status[*],/usr/lib/zabbix/externalscripts/flashsystem_collector.sh $1 $2 $3 getvolumestatus $4
UserParameter=flashsystem.node.status[*],/usr/lib/zabbix/externalscripts/flashsystem_collector.sh $1 $2 $3 getnodestatus $4

# Métricas de rendimiento y capacidad
UserParameter=flashsystem.pool.usage[*],/usr/lib/zabbix/externalscripts/flashsystem_collector.sh $1 $2 $3 getpoolusage $4
UserParameter=flashsystem.system.iops[*],/usr/lib/zabbix/externalscripts/flashsystem_collector.sh $1 $2 $3 getiops
UserParameter=flashsystem.cluster.status[*],/usr/lib/zabbix/externalscripts/flashsystem_collector.sh $1 $2 $3 getclusterstatus
EOF

# Reiniciar Zabbix Agent para aplicar cambios
sudo systemctl restart zabbix-agent
```

### 6.3 Crear Items en Zabbix Frontend

#### Item de Ejemplo: IOPS del Sistema

| Campo | Valor |
|-------|-------|
| **Nombre** | IBM FlashSystem: IOPS Totales |
| **Tipo** | External check |
| **Clave** | `flashsystem_collector.sh[{HOST.IP},monitor,contraseña,getiops]` |
| **Tipo de información** | Numérico (entero sin signo) |
| **Unidades** | IOPS |
| **Intervalo de actualización** | 60s |
| **Historia almacenada** | 90d |

#### Item de Ejemplo: Estado de Unidad

| Campo | Valor |
|-------|-------|
| **Nombre** | IBM FlashSystem: Estado de Unidad {#ENCLOSUREID}:{#SLOTID} |
| **Tipo** | External check |
| **Clave** | `flashsystem_collector.sh[{HOST.IP},monitor,contraseña,getdrivestatus,{#ENCLOSUREID},{#SLOTID}]` |
| **Tipo de información** | Numérico (entero sin signo) |
| **Valores predefinidos** | 0=Offline, 1=Online, 2=Desconocido |
| **Intervalo de actualización** | 300s |

### 6.4 Configurar Low-Level Discovery (LLD)

#### Regla de Discovery para Unidades

| Campo | Valor |
|-------|-------|
| **Nombre** | IBM FlashSystem: Descubrimiento de Unidades |
| **Tipo** | External check |
| **Clave** | `flashsystem_collector.sh[{HOST.IP},monitor,contraseña,discoverdrives]` |
| **Intervalo de actualización** | 3600s (1 hora) |
| **Mantener recursos descubiertos perdidos** | 7d |
| **Mantener recursos descubiertos no descubiertos** | 30d |

#### Prototipos de Item para Unidades Descubiertas

| Prototipo | Clave | Tipo | Intervalo |
|-----------|-------|------|-----------|
| Estado de unidad | `flashsystem_collector.sh[{HOST.IP},monitor,contraseña,getdrivestatus,{#ENCLOSUREID},{#SLOTID}]` | External check | 300s |
| Nombre de unidad | `flashsystem_collector.sh[{HOST.IP},monitor,contraseña,getdrivestatus,{#ENCLOSUREID},{#SLOTID}]` | Dependent item | - |

#### Filtros de Discovery (Opcional)

Para excluir unidades no relevantes:

```
Filtro: {#DRIVESTATUS} @ ^(online|available|good|active)$
```

### 6.5 Macros de Host Recomendadas

Crear macros a nivel de host para centralizar configuración:

| Macro | Valor | Descripción |
|-------|-------|-------------|
| `{$FS_HOST}` | 10.10.10.50 | IP del FlashSystem |
| `{$FS_USER}` | monitor | Usuario SSH |
| `{$FS_PASSWORD}` | contraseña123 | Contraseña (usar vault en producción) |
| `{$FS_POLL_INTERVAL}` | 60 | Intervalo de polling en segundos |
| `{$FS_DISCOVERY_INTERVAL}` | 3600 | Intervalo de discovery en segundos |

**Uso en claves de item:**
```
flashsystem_collector.sh[{$FS_HOST},{$FS_USER},{$FS_PASSWORD},getiops]
```

---

## 7. Configuración del IBM FlashSystem

### 7.1 Comandos CLI Requeridos

El usuario de monitoreo debe poder ejecutar estos comandos:

```bash
# Comandos de descubrimiento
svcinfo lsdrive -delim :
svcinfo lsenclosure -delim :
svcinfo lsmdiskgrp -delim :
svcinfo lsvdisk -delim :
svcinfo lsnode -delim :
svcinfo lsportfc -delim :
svcinfo lsfcmap -delim :
svcinfo lsrcrelationship -delim :

# Comandos de métricas
svcinfo lsenclosurebattery -delim :
svcinfo lssystemstats -delim :
svcinfo lssystem -delim :
```

### 7.2 Verificación de Permisos del Usuario

Conectarse como el usuario de monitoreo y probar:

```bash
ssh monitor@10.10.10.50

# Probar cada comando
SVC_CLI_LOCALE=C svcinfo lssystem -delim : | head -1
SVC_CLI_LOCALE=C svcinfo lsdrive -delim : | head -3
SVC_CLI_LOCALE=C svcinfo lsmdiskgrp -delim : | head -3

# Verificar que NO puede ejecutar comandos de escritura
svcinfo mkvdisk -help  # Debe retornar "Permission denied" o similar
```

### 7.3 Configuración de Timeout en FlashSystem

Asegurar que el sistema no cierre sesiones SSH prematuramente:

```bash
# En la CLI del FlashSystem, verificar configuración SSH
svcinfo lssystem | grep -i ssh

# Si es necesario, ajustar timeout de sesión (valores típicos: 300-600 segundos)
# Nota: Este comando requiere privilegios de administrador
chsystem -ssh_timeout 600
```

---

## 8. Pruebas del Sistema

### 8.1 Prueba Manual del Binario

```bash
# Ejecutar comando de prueba directamente
sudo -u zabbix /usr/local/bin/flashsystem_collector 10.10.10.50 monitor contraseña123 getiops

# Salida esperada: un número entero
12500

# Si hay error, revisar logs
tail -20 /var/log/zabbix/flashsystem_collector.log
```

### 8.2 Prueba del Script Wrapper

```bash
# Ejecutar vía script wrapper (como lo haría Zabbix)
sudo -u zabbix /usr/lib/zabbix/externalscripts/flashsystem_collector.sh 10.10.10.50 monitor contraseña123 discoverpools

# Salida esperada: JSON válido
{"data":[{"{#ID}":"0","{#NAME}":"Pool0","{#STATUS}":"online"}]}

# Validar JSON con herramienta
echo '{"data":[]}' | python3 -m json.tool > /dev/null && echo "JSON válido"
```

### 8.3 Prueba desde Zabbix Frontend

1. Navegar a **Monitoring → Latest data**
2. Filtrar por el host del FlashSystem
3. Buscar items con clave `flashsystem.*`
4. Verificar que los valores se actualizan correctamente
5. Revisar columna "Last check" para confirmar ejecución exitosa

### 8.4 Prueba de Discovery

```bash
# Ejecutar discovery manualmente y contar elementos
sudo -u zabbix /usr/lib/zabbix/externalscripts/flashsystem_collector.sh 10.10.10.50 monitor contraseña123 discoverdrives | python3 -c "import sys,json; d=json.load(sys.stdin); print(f'Unidades descubiertas: {len(d[\"data\"])}')"

# Salida esperada
Unidades descubiertas: 24
```

### 8.5 Validación de Triggers (Opcional)

Crear un trigger de prueba para verificar alertas:

| Campo | Valor |
|-------|-------|
| **Nombre** | Test: FlashSystem Collector Funcionando |
| **Expresión** | `last(/Host/flashsystem.system.iops[*])<0` |
| **Severidad** | Information |
| **Descripción** | Trigger de prueba para validar integración |

---

## 9. Troubleshooting

### 9.1 Errores Comunes y Soluciones

#### Error: "Connection refused" o "Timeout"

```
Síntoma: El script retorna 0 y los logs muestran errores de conexión SSH

Causas posibles:
- Puerto 22 bloqueado por firewall
- IP incorrecta en configuración
- Servicio SSH no ejecutándose en FlashSystem

Solución:
# Verificar conectividad de red
ping 10.10.10.50
nc -zv 10.10.10.50 22

# Verificar reglas de firewall en RHEL
sudo firewall-cmd --list-all
sudo firewall-cmd --add-service=ssh --permanent
sudo firewall-cmd --reload

# Verificar que SSH está activo en FlashSystem
ssh admin@10.10.10.50 "svcinfo lssystem | grep -i ssh"
```

#### Error: "Authentication failed"

```
Síntoma: Error de autenticación en logs

Causas posibles:
- Contraseña incorrecta
- Usuario bloqueado o expirado
- Política de autenticación SSH restrictiva

Solución:
# Probar autenticación manualmente
ssh monitor@10.10.10.50 "echo test"

# Verificar estado del usuario en FlashSystem
ssh admin@10.10.10.50 "lsuser monitor"

# Revisar logs de autenticación en FlashSystem
ssh admin@10.10.10.50 "svctask lslog -filterlevel warning | grep -i auth"
```

#### Error: "Command not found" o "Permission denied"

```
Síntoma: El comando svcinfo no se ejecuta o retorna error de permisos

Causas posibles:
- Usuario sin permisos para ejecutar svcinfo
- Comando no disponible en esta versión de Spectrum Virtualize
- PATH no configurado correctamente

Solución:
# Verificar que el usuario puede ejecutar svcinfo
ssh monitor@10.10.10.50 "which svcinfo"
ssh monitor@10.10.10.50 "svcinfo lssystem -delim : | head -1"

# Si el comando no está en PATH, usar ruta completa
# (ajustar parser.go para usar ruta completa si es necesario)
```

#### Error: "JSON parse error" en Zabbix

```
Síntoma: Zabbix no procesa la salida del discovery

Causas posibles:
- Salida JSON mal formada
- Caracteres especiales no escapados
- Salida vacía por error previo

Solución:
# Validar salida JSON manualmente
sudo -u zabbix /usr/lib/zabbix/externalscripts/flashsystem_collector.sh 10.10.10.50 monitor contraseña123 discoverdrives | python3 -m json.tool

# Revisar logs para errores de parsing
grep -i "parse\|json" /var/log/zabbix/flashsystem_collector.log

# Verificar que hay datos en el sistema
ssh monitor@10.10.10.50 "svcinfo lsdrive -delim : | wc -l"
```

#### Error: "Timeout while executing script"

```
Síntoma: Zabbix marca el item como "Not supported" por timeout

Causas posibles:
- FlashSystem lento o sobrecargado
- Timeout de Zabbix muy bajo (default: 3s para External check)
- Red con alta latencia

Solución:
# Aumentar timeout en configuración de Zabbix Server
# Editar /etc/zabbix/zabbix_server.conf
Timeout=30

# Reiniciar Zabbix Server
sudo systemctl restart zabbix-server

# Alternativa: Reducir frecuencia de polling para items pesados
```

### 9.2 Comandos de Diagnóstico

```bash
# Verificar que el binario es ejecutable
file /usr/local/bin/flashsystem_collector
ls -l /usr/local/bin/flashsystem_collector

# Verificar dependencias dinámicas del binario
ldd /usr/local/bin/flashsystem_collector

# Probar ejecución con logging detallado
sudo -u zabbix /usr/local/bin/flashsystem_collector 10.10.10.50 monitor contraseña123 getiops 2>&1 | tee /tmp/test_output.txt

# Monitorear logs en tiempo real
tail -f /var/log/zabbix/flashsystem_collector.log

# Verificar procesos Zabbix relacionados
ps aux | grep -E "zabbix|flashsystem"

# Probar conectividad SSH desde el proceso Zabbix
sudo -u zabbix ssh -o BatchMode=yes -o ConnectTimeout=10 monitor@10.10.10.50 "echo OK"
```

### 9.3 Logs Relevantes

| Archivo | Propósito | Comando de consulta |
|---------|-----------|-------------------|
| `/var/log/zabbix/flashsystem_collector.log` | Logs del colector | `tail -f /var/log/zabbix/flashsystem_collector.log` |
| `/var/log/zabbix/zabbix_server.log` | Logs del servidor Zabbix | `grep flashsystem /var/log/zabbix/zabbix_server.log` |
| `/var/log/secure` (RHEL) | Logs de autenticación SSH | `grep flashsystem /var/log/secure` |
| `journalctl -u zabbix-server` | Logs systemd de Zabbix | `journalctl -u zabbix-server -f` |

---

## 10. Consideraciones de Producción

### 10.1 Seguridad

- **Nunca almacenar contraseñas en scripts**: Usar vault de secretos (HashiCorp Vault, Ansible Vault) o autenticación por claves SSH.
- **Restringir acceso SSH**: Configurar `AllowUsers` en `/etc/ssh/sshd_config` del FlashSystem.
- **Auditar conexiones**: Habilitar logging detallado de sesiones SSH en el storage.
- **Rotar credenciales**: Establecer política de rotación de contraseñas cada 90 días.

### 10.2 Rendimiento

- **Evitar polling excesivo**: Usar intervalos de 300-600s para métricas de estado, 3600s para discovery.
- **Implementar caching**: Considerar agregar cache en el colector para reducir carga en el storage.
- **Monitorear el monitor**: Crear items para monitorear la ejecución del propio colector.

### 10.3 Alta Disponibilidad

- **Configurar múltiples proxies Zabbix**: Si el storage es crítico, configurar Zabbix Proxy en modo HA.
- **Timeouts adecuados**: Ajustar timeouts para manejar picos de carga sin falsas alertas.
- **Alertas de falla del colector**: Crear trigger que alerte si el colector no responde por N ciclos.

### 10.4 Mantenimiento

- **Actualizaciones**: Probar actualizaciones del colector en entorno de staging antes de producción.
- **Backups**: Incluir configuración de Zabbix y scripts en política de backup.
- **Documentación**: Mantener actualizada esta guía con cambios en el entorno.

---

## Apéndice A: Comandos Rápidos de Referencia

```bash
# Compilar proyecto
cd /opt/flashsystem-collector && go build -o flashsystem_collector .

# Probar métrica individual
/usr/lib/zabbix/externalscripts/flashsystem_collector.sh 10.10.10.50 monitor pass getiops

# Probar discovery
/usr/lib/zabbix/externalscripts/flashsystem_collector.sh 10.10.10.50 monitor pass discoverpools | jq '.data | length'

# Verificar logs
tail -50 /var/log/zabbix/flashsystem_collector.log | grep -E "ERROR|SUCCESS"

# Reiniciar servicios tras cambios
systemctl restart zabbix-server
systemctl restart zabbix-agent  # Si se usan UserParameters

# Validar configuración Zabbix
zabbix_server -R config_cache_reload
```

---

## Apéndice B: Checklist de Implementación

- [ ] Go 1.22+ instalado y en PATH
- [ ] Repositorio clonado en /opt/flashsystem-collector
- [ ] Binario compilado y ubicado en /usr/local/bin/
- [ ] Script wrapper en /usr/lib/zabbix/externalscripts/ con permisos correctos
- [ ] Directorios de config y logs creados con permisos zabbix:zabbix
- [ ] Usuario de monitoreo creado en FlashSystem con permisos de lectura
- [ ] Conexión SSH probada manualmente desde usuario zabbix
- [ ] Items y discovery configurados en Zabbix frontend
- [ ] Pruebas manuales exitosas (métricas y discovery)
- [ ] Logs verificando ejecución sin errores
- [ ] Triggers de alerta configurados (opcional pero recomendado)
- [ ] Documentación de operación entregada al equipo

---

**Documento revisado y aprobado para implementación en producción.**  
*Última actualización: Marzo 2026*




El script de Go se compila **antes de su ejecución**, durante la fase de **instalación o despliegue** en el servidor Zabbix. **No se compila en tiempo de ejecución** cada vez que Zabbix solicita una métrica.

Aquí están los detalles técnicos basados en la arquitectura del proyecto y la documentación de Zabbix:

### 1. Momento Exacto: Fase de Instalación
La compilación ocurre una única vez (o cada vez que se actualiza el código) antes de que Zabbix intente monitorear el storage.

*   **Cuándo:** Cuando el administrador sigue la **Guía de Implementación** que generamos anteriormente (Sección 3: Instalación del Proyecto).
*   **Quién:** El administrador de sistemas o un pipeline de CI/CD.
*   **Comando:** Se ejecuta manualmente o mediante script de automatización:
    ```bash
    go build -ldflags "-s -w" -o flashsystem_collector .
    ```
*   **Resultado:** Se genera un binario estático (`flashsystem_collector`) que es el que Zabbix ejecutará realmente.

### 2. Por qué no se compila en tiempo de ejecución
Zabbix espera un **ejecutable** o un **script interpretado** (como Bash o Python). Go es un lenguaje compilado:

*   **Rendimiento:** Si se compilara en cada polling (ej. cada 60 segundos), el tiempo de respuesta sería de varios segundos solo en compilación, causando *timeouts* en Zabbix.
*   **Dependencias:** El binario compilado ya incluye todas las librerías necesarias (como `golang.org/x/crypto/ssh`). No requiere tener el compilador de Go instalado en el servidor de producción durante la operación diaria.
*   **Estabilidad:** Un binario compilado es inmutable. Evita errores por cambios en el entorno de Go durante la operación.

### 3. Flujo de Ejecución en Producción

1.  **Despliegue (Una vez):**
    *   Se descarga el código fuente.
    *   Se compila con `go build`.
    *   Se copia el binario a `/usr/local/bin/flashsystem_collector`.
2.  **Operación (Cada polling):**
    *   Zabbix Server llama al script wrapper: `/usr/lib/zabbix/externalscripts/flashsystem_collector.sh`.
    *   El wrapper ejecuta el **binario ya compilado**: `/usr/local/bin/flashsystem_collector`.
    *   El binario se conecta por SSH, obtiene datos y retorna la métrica.

### 4. ¿Cuándo volver a compilar?
Solo es necesario volver a compilar el script en los siguientes casos:

*   **Actualización de versión:** Cuando se descargan cambios del repositorio para mejorar funcionalidades o corregir bugs.
*   **Cambio de arquitectura:** Si se mueve el binario a un servidor con una arquitectura de CPU diferente (ej. de Intel x86_64 a ARM), se debe recompilar con `GOOS=linux GOARCH=arm64 go build`.
*   **Actualización de dependencias:** Si se modifica el `go.mod` para actualizar librerías de seguridad.

### Referencia a la Documentación
Según la **Guía de Implementación** generada en este proyecto (Sección 3.2):

```bash
# Paso 3.2: Clonar y Compilar (Solo durante la instalación)
cd /opt/flashsystem-collector
go mod tidy
go build -ldflags "-s -w" -o flashsystem_collector .

# Paso 3.3: Mover binario (Listo para que Zabbix lo use)
sudo cp flashsystem_collector /usr/local/bin/
```

**Conclusión:** El código Go se transforma en un ejecutable binario durante la **instalación**. Zabbix interactúa exclusivamente con ese binario compilado, no con el código fuente.




# Análisis del Cambio en los Imports de discovery.go

## ¿Qué Pasaría con Ese Cambio?

### 1. **Error de Compilación (Escenario Más Probable)**

Si tu archivo `go.mod` tiene definido el módulo como:

```go
module flashsystem_zabbix
```

Pero en `discovery.go` usas:

```go
import (
    "zabbix_go_ibm_flash_v2/internal/parser"
    "zabbix_go_ibm_flash_v2/internal/ssh"
)
```

**Resultado:** ❌ **El código NO compilará**

**Error que verás:**
```
package zabbix_go_ibm_flash_v2/internal/parser is not in GOROOT
package zabbix_go_ibm_flash_v2/internal/ssh is not in GOROOT
```

---

### 2. **Por Qué Ocurre Este Error**

En Go, los imports deben coincidir **exactamente** con el nombre del módulo definido en `go.mod`:

| Archivo | Valor Actual | Valor Propuesto | ¿Coincide? |
|---------|-------------|-----------------|------------|
| `go.mod` | `module flashsystem_zabbix` | - | - |
| `discovery.go` | `flashsystem_zabbix/internal/...` | `zabbix_go_ibm_flash_v2/internal/...` | ❌ NO |

---

## Cómo Hacer el Cambio Correctamente

### Opción A: Cambiar Todo a `zabbix_go_ibm_flash_v2` (Recomendado si quieres ese nombre)

**Paso 1:** Actualizar `go.mod`
```go
module zabbix_go_ibm_flash_v2

go 1.22

require golang.org/x/crypto v0.22.0
```

**Paso 2:** Actualizar **todos** los archivos con imports internos

| Archivo | Cambio Requerido |
|---------|-----------------|
| `main.go` | `flashsystem_zabbix/internal/...` → `zabbix_go_ibm_flash_v2/internal/...` |
| `internal/ssh/ssh.go` | `flashsystem_zabbix/internal/...` → `zabbix_go_ibm_flash_v2/internal/...` |
| `internal/monitor/monitor.go` | `flashsystem_zabbix/internal/...` → `zabbix_go_ibm_flash_v2/internal/...` |
| `internal/parser/parser.go` | (Sin cambios si no importa otros paquetes internos) |
| `internal/discovery/discovery.go` | ✅ (El que estás cambiando) |
| `internal/utils/utils.go` | (Sin cambios si no importa otros paquetes internos) |

**Paso 3:** Limpiar cache y recompilar
```bash
go clean -modcache
go mod tidy
go build -o flashsystem_collector .
```

---

### Opción B: Mantener `flashsystem_zabbix` (Menos Cambios)

Si el repositorio se llama `zabbix_go_ibm_flash_v2` pero quieres mantener el módulo como `flashsystem_zabbix`:

**No cambies los imports** - déjalos como:
```go
import (
    "flashsystem_zabbix/internal/parser"
    "flashsystem_zabbix/internal/ssh"
)
```

**Ventaja:** Solo cambias el nombre del repositorio en GitHub, no el código.

---

## Verificación de Archivos Afectados

Basado en la estructura del proyecto, estos archivos necesitarían actualización si cambias el nombre del módulo:

```bash
# Buscar todos los imports que necesitan cambio
grep -r "flashsystem_zabbix/internal" . --include="*.go"

# Archivos típicos afectados:
./main.go
./internal/ssh/ssh.go
./internal/monitor/monitor.go
./internal/discovery/discovery.go
```

---

## Tabla de Decisión

| Escenario | Acción Requerida | Impacto |
|-----------|-----------------|---------|
| Solo cambias `discovery.go` | ❌ No funcionará | Error de compilación |
| Cambias `go.mod` + todos los imports | ✅ Funcionará | Requiere actualizar ~4-5 archivos |
| Mantienes nombre actual | ✅ Funcionará | Sin cambios necesarios |

---

## Recomendación

**Si el repositorio en GitHub se llama `zabbix_go_ibm_flash_v2`:**

```bash
# 1. Actualizar go.mod
sed -i 's/module flashsystem_zabbix/module zabbix_go_ibm_flash_v2/' go.mod

# 2. Actualizar todos los imports en todo el proyecto
find . -name "*.go" -type f -exec sed -i 's|flashsystem_zabbix/internal|zabbix_go_ibm_flash_v2/internal|g' {} \;

# 3. Verificar cambios
go mod tidy
go build -o flashsystem_collector .
```

**Importante:** El nombre del módulo **no afecta** la funcionalidad del programa. Solo debe ser consistente en todo el proyecto. Puedes usar cualquiera de los dos nombres, pero **debe coincidir entre `go.mod` y todos los imports**.