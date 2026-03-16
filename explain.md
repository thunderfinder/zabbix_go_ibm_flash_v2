# Explicación Técnica del Proyecto IBM FlashSystem Zabbix Collector

## 1. Propósito del proyecto

Este proyecto resuelve el problema de monitorear sistemas de almacenamiento IBM FlashSystem e IBM Spectrum Virtualize desde Zabbix. Sin esta herramienta, los administradores tendrían que:

* Conectarse manualmente al storage para verificar estado
* No tendrían alertas automáticas cuando hay problemas
* No podrían ver tendencias históricas de rendimiento

El colector actúa como un **puente** entre el sistema de almacenamiento y Zabbix, permitiendo:

* Monitoreo automático de componentes críticos (discos, carcasas, pools, volúmenes)
* Alertas proactivas cuando algo falla
* Histórico de métricas para análisis de capacidad y rendimiento

---

## 2. Flujo general del sistema

El proceso completo funciona en estos pasos:

```
Paso 1: Zabbix solicita una métrica
   ↓
Paso 2: Zabbix ejecuta el script flashsystem_collector.sh
   ↓
Paso 3: El script llama al binario Go (flashsystem_collector)
   ↓
Paso 4: El programa Go se conecta por SSH al IBM FlashSystem
   ↓
Paso 5: Ejecuta comandos svcinfo en el storage
   ↓
Paso 6: Recibe la salida de texto de los comandos
   ↓
Paso 7: Analiza y convierte la salida a formato estructurado
   ↓
Paso 8: Devuelve el resultado a Zabbix (JSON o valor simple)
   ↓
Paso 9: Zabbix almacena la métrica y evalúa triggers
```

**Tiempo típico de ejecución:** 2-10 segundos por comando, dependiendo de la cantidad de datos.

---

## 3. Funcionamiento de main.go

El archivo `main.go` es el **punto de entrada** del programa. Cuando se ejecuta, hace lo siguiente:

### 3.1 Validación de argumentos

El programa espera al menos 4 argumentos en este orden:

| Posición | Contenido | Ejemplo |
|----------|-----------|---------|
| 1 | Dirección IP del FlashSystem | 10.10.10.50 |
| 2 | Usuario SSH | monitor |
| 3 | Contraseña SSH | contraseña123 |
| 4 | Comando a ejecutar | getdrivestatus |
| 5+ | Parámetros adicionales | 1 12 (enclosure y slot) |

Si faltan argumentos, el programa imprime `0` y termina con error.

### 3.2 Validación del comando

El programa verifica que el comando esté en una **lista blanca** de comandos permitidos. Esto es una medida de seguridad para prevenir ejecución de comandos no autorizados.

Comandos válidos:
* `discoverdrives`, `discoverenclosures`, `discoverpools`, `discovervolumes`
* `getdrivestatus`, `getenclosurestatus`, `getbatterystatus`
* `getpoolusage`, `getiops`, `getvolumestatus`, `getnodestatus`, `getclusterstatus`

### 3.3 Ejecución del comando

Según el comando recibido, `main.go` llama a la función correspondiente:

* Para **discovery**: llama a funciones del paquete `discovery`
* Para **métricas**: llama a funciones del paquete `monitor`

---

## 4. Conexión con el sistema IBM FlashSystem

### 4.1 Protocolo de comunicación

El programa se conecta al IBM FlashSystem usando **SSH (puerto 22)**. Esta es la misma forma en que un administrador se conectaría manualmente, pero automatizada.

### 4.2 Configuración de la conexión SSH

La conexión se configura con estos parámetros:

| Parámetro | Valor | Propósito |
|-----------|-------|-----------|
| Timeout de conexión | 10 segundos | Evita esperas infinitas si el sistema no responde |
| Timeout de comando | 30 segundos | Limita el tiempo de ejecución de cada comando |
| Locale | SVC_CLI_LOCALE=C | Fuerza salida en inglés para parsing consistente |
| Verificación de host key | Deshabilitada | Permite conexión sin configuración previa (cambiar en producción) |

### 4.3 Gestión de sesiones

Cada comando SSH:
1. Crea una **nueva sesión** SSH
2. Ejecuta el comando
3. Captura la salida estándar
4. Cierra la sesión inmediatamente

Esto asegura que no queden sesiones abiertas que puedan agotar recursos.

### 4.4 Reintentos automáticos

Si un comando falla, el programa intenta **hasta 3 veces** con espera progresiva:

* Intento 1: Falla → espera 1 segundo
* Intento 2: Falla → espera 2 segundos
* Intento 3: Falla → reporta error definitivo

Esto maneja problemas temporales de red o carga del sistema.

---

## 5. Obtención de datos del storage

### 5.1 Comandos CLI ejecutados

El programa ejecuta comandos `svcinfo` que son parte de la interfaz de línea de comandos de IBM Spectrum Virtualize. Estos son los comandos principales:

| Comando | Propósito | Información obtenida |
|---------|-----------|---------------------|
| `svcinfo lsdrive -delim :` | Listar unidades físicas | ID, enclosure, slot, estado, tipo |
| `svcinfo lsenclosure -delim :` | Listar carcasas | ID, nombre, estado, temperatura |
| `svcinfo lsenclosurebattery -delim :` | Listar baterías | ID, enclosure, estado de carga |
| `svcinfo lsmdiskgrp -delim :` | Listar grupos de discos | Nombre, capacidad total, capacidad libre |
| `svcinfo lsvdisk -delim :` | Listar volúmenes virtuales | Nombre, ID, estado, tipo |
| `svcinfo lsnode -delim :` | Listar nodos del cluster | ID, nombre, estado, tipo |
| `svcinfo lsportfc -delim :` | Listar puertos Fibre Channel | ID, WWPN, estado, velocidad |
| `svcinfo lsfcmap -delim :` | Listar relaciones FlashCopy | ID, origen, destino, estado |
| `svcinfo lsrcrelationship -delim :` | Listar replicación remota | ID, cluster maestro, cluster auxiliar |
| `svcinfo lssystemstats -delim :` | Estadísticas del sistema | IOPS de lectura/escritura |
| `svcinfo lssystem -delim :` | Información del sistema | Nombre, versión, estado del cluster |

### 5.2 Formato de salida

Los comandos retornan texto con este formato:

```
id:name:status:capacity:free_capacity
0:Pool0:online:1099511627776:549755813888
1:Pool1:online:2199023255552:1099511627776
```

* Primera línea: **encabezados** de columna separados por `:`
* Líneas siguientes: **valores** separados por `:`
* El parámetro `-delim :` asegura que el separador sea dos puntos

### 5.3 Configuración de locale

Antes de cada comando, el programa establece:

```
SVC_CLI_LOCALE=C
```

Esto es **crítico** porque:
* Asegura que los nombres de campos estén en inglés
* Evita problemas con formatos de número locales (comas vs puntos)
* Garantiza que el parsing funcione consistentemente

---

## 6. Procesamiento de datos

### 6.1 Análisis de la salida (Parsing)

El paquete `parser` convierte la salida de texto en estructuras de datos Go. El proceso es:

1. **Dividir por líneas**: Separa la salida completa en líneas individuales
2. **Extraer encabezados**: La primera línea contiene los nombres de los campos
3. **Propiar cada línea de datos**: Divide por `:` y asocia cada valor con su encabezado
4. **Crear registros**: Cada fila se convierte en un mapa `campo → valor`

**Ejemplo de transformación:**

```
Salida original (texto):
id:name:status
1:Drive1:online
2:Drive2:offline

Resultado después del parsing:
[
  {"id": "1", "name": "Drive1", "status": "online"},
  {"id": "2", "name": "Drive2", "status": "offline"}
]
```

### 6.2 Limpieza de datos

El parser también:
* Elimina espacios en blanco al inicio y final de cada valor
* Remueve caracteres de control que podrían causar problemas
* Filtra líneas vacías
* Maneja diferentes formatos de salto de línea (Windows/Unix)

### 6.3 Conversión de valores numéricos

Para métricas como uso de pool o IOPS:
* Convierte cadenas a números (ej: `"1024"` → `1024`)
* Maneja unidades como GB, TB, MB si están presentes
* Valida que los valores sean razonables (ej: libre no mayor que total)

---

## 7. Integración con Zabbix

### 7.1 Formato de salida

El programa devuelve datos en dos formatos según el tipo de comando:

**Para métricas individuales:**
```
1          ← Estado online (número simple)
0          ← Estado offline
85.50      ← Porcentaje de uso
12500      ← IOPS totales
```

Zabbix lee este valor directamente y lo almacena como métrica.

**Para descubrimiento (LLD):**
```json
{"data":[
  {"{#DRIVEID}":"1","{#ENCLOSUREID}":"0","{#SLOTID}":"12","{#DRIVESTATUS}":"online"},
  {"{#DRIVEID}":"2","{#ENCLOSUREID}":"0","{#SLOTID}":"13","{#DRIVESTATUS}":"online"}
]}
```

Zabbix usa este JSON para crear automáticamente items para cada unidad descubierta.

### 7.2 Códigos de estado

Los estados se convierten a números para Zabbix:

| Estado | Valor | Significado |
|--------|-------|-------------|
| online/available/good | 1 | OK - Todo funciona |
| offline/failed/error | 0 | Problema - Requiere atención |
| otros estados | 2 | Desconocido - Verificar manualmente |

### 7.3 Tipos de items en Zabbix

| Tipo de comando | Tipo de item Zabbix | Ejemplo de key |
|-----------------|---------------------|----------------|
| Estado de unidad | Numérico (entero) | `flashsystem.drive.status[1,12]` |
| Uso de pool | Numérico (float) | `flashsystem.pool.usage[Pool0]` |
| IOPS | Numérico (entero) | `flashsystem.system.iops[]` |
| Discovery | Texto | `flashsystem.discovery.drives[]` |

---

## 8. Descubrimiento automático (LLD)

### 8.1 Qué es Low-Level Discovery

LLD permite que Zabbix **descubra automáticamente** componentes del storage sin configuración manual. Cuando aparece un nuevo disco, Zabbix lo detecta y crea items para él automáticamente.

### 8.2 Cómo funciona en este proyecto

1. Zabbix ejecuta el comando de discovery (ej: `discoverdrives`)
2. El programa consulta al storage todas las unidades
3. Devuelve un JSON con la lista de unidades
4. Zabbix crea un item para cada unidad descubierta
5. En cada ciclo de polling, Zabbix verifica el estado de cada unidad

### 8.3 Macros de descubrimiento

Cada tipo de discovery define macros específicas:

**Para discos:**
* `{#DRIVEID}` - ID único del disco
* `{#ENCLOSUREID}` - ID del enclosure donde está
* `{#SLOTID}` - Número de slot
* `{#DRIVESTATUS}` - Estado actual

**Para pools:**
* `{#ID}` - ID del pool
* `{#NAME}` - Nombre del pool
* `{#STATUS}` - Estado del pool

**Para volúmenes:**
* `{#ID}` - ID del volumen
* `{#NAME}` - Nombre del volumen
* `{#STATUS}` - Estado del volumen
* `{#VDISK_TYPE}` - Tipo de volumen

### 8.4 Prototipos de items

Con el discovery, se pueden crear prototipos como:

* **Item**: Estado del disco `{#DRIVEID}` → Key: `getdrivestatus[{#ENCLOSUREID},{#SLOTID}]`
* **Trigger**: Disco offline → `{last()}=0`
* **Gráfico**: Uso de pool por nombre → `{#NAME}`

---

## 9. Manejo de errores

### 9.1 Tipos de errores manejados

| Tipo de error | Comportamiento del programa | Valor devuelto a Zabbix |
|---------------|----------------------------|------------------------|
| Conexión SSH fallida | Reintenta 3 veces, luego falla | `0` |
| Comando timeout (30s) | Cierra sesión, reporta error | `0` |
| Comando no encontrado | Registra error, sale | `0` |
| Parsing fallido | Registra error, devuelve vacío | `{}` o `0` |
| Argumentos inválidos | Valida antes de ejecutar | `0` |
| Recurso no encontrado | Busca en resultados, no encuentra | `0` |

### 9.2 Logging de errores

Todos los errores se registran en:
```
/var/log/zabbix/flashsystem_collector.log
```

Cada entrada incluye:
* Fecha y hora
* Nivel de severidad (ERROR, WARNING, INFO)
* Nombre del archivo y línea donde ocurrió
* Mensaje descriptivo del error

### 9.3 Comportamiento ante fallos

El programa sigue el principio de **"fallar suavemente"**:

* Nunca termina abruptamente sin aviso
* Siempre devuelve un valor que Zabbix puede procesar
* Registra el error para diagnóstico posterior
* Permite que Zabbix marque el item como "no disponible" si hay problemas persistentes

---

## 10. Estructura del proyecto

```
zabbix_go_ibm_flash_v2/
├── go.mod                          # Dependencias del proyecto
├── main.go                         # Punto de entrada principal
├── internal/
│   ├── ssh/
│   │   └── ssh.go                  # Conexión y comandos SSH
│   ├── parser/
│   │   └── parser.go               # Análisis de salida de comandos
│   ├── monitor/
│   │   └── monitor.go              # Funciones de monitoreo específicas
│   ├── discovery/
│   │   └── discovery.go            # Funciones de descubrimiento LLD
│   └── utils/
│       └── utils.go                # Funciones de utilidad general
├── scripts/
│   └── flashsystem_collector.sh    # Wrapper para Zabbix External Scripts
└── README.md                       # Documentación del proyecto
```

### Propósito de cada archivo:

| Archivo | Responsabilidad |
|---------|-----------------|
| `go.mod` | Define el módulo Go y dependencias externas (golang.org/x/crypto para SSH) |
| `main.go` | Recibe argumentos, valida comandos, coordina la ejecución |
| `ssh.go` | Establece conexiones SSH, ejecuta comandos remotos, maneja timeouts |
| `parser.go` | Convierte salida de texto en estructuras de datos, limpia valores |
| `monitor.go` | Implementa lógica para cada tipo de métrica (estado, uso, IOPS) |
| `discovery.go` | Genera JSON de descubrimiento para LLD de Zabbix |
| `utils.go` | Funciones auxiliares (validación, conversión, formateo) |
| `flashsystem_collector.sh` | Script bash que Zabbix ejecuta, maneja logging y permisos |

---

## 11. Ejemplo de ejecución real

### 11.1 Ejecución manual para pruebas

```bash
# Obtener estado de un disco específico (enclosure 1, slot 12)
/usr/lib/zabbix/externalscripts/flashsystem_collector.sh \
  10.10.10.50 monitor contraseña123 getdrivestatus 1 12

# Salida esperada: 1 (online) o 0 (offline)
```

```bash
# Descubrir todos los grupos de discos
/usr/lib/zabbix/externalscripts/flashsystem_collector.sh \
  10.10.10.50 monitor contraseña123 discoverpools

# Salida esperada: JSON con lista de pools
{"data":[{"{#ID}":"0","{#NAME}":"Pool0","{#STATUS}":"online"}]}
```

```bash
# Obtener porcentaje de uso de un pool
/usr/lib/zabbix/externalscripts/flashsystem_collector.sh \
  10.10.10.50 monitor contraseña123 getpoolusage Pool0

# Salida esperada: 45.50 (45.5% de uso)
```

```bash
# Obtener IOPS totales del sistema
/usr/lib/zabbix/externalscripts/flashsystem_collector.sh \
  10.10.10.50 monitor contraseña123 getiops

# Salida esperada: 15000 (IOPS combinadas de lectura + escritura)
```

### 11.2 Configuración en Zabbix

**Item para estado de disco:**
* Tipo: External check
* Key: `flashsystem_collector.sh[{HOST.IP},monitor,contraseña,getdrivestatus,1,12]`
* Tipo de información: Numérico (entero sin signo)
* Intervalo de actualización: 60s

**Regla de discovery para discos:**
* Tipo: External check
* Key: `flashsystem_collector.sh[{HOST.IP},monitor,contraseña,discoverdrives]`
* Intervalo de actualización: 3600s (una vez por hora es suficiente)

### 11.3 Verificación de logs

```bash
# Ver logs del colector
tail -f /var/log/zabbix/flashsystem_collector.log

# Ejemplo de entrada de log:
[2025-01-15 10:30:45] [flashsystem_collector] Iniciando ejecución con argumentos: 10.10.10.50 *** *** getdrivestatus
[2025-01-15 10:30:47] [flashsystem_collector] SUCCESS: Colector terminó exitosamente
```

---

## Resumen ejecutivo

Este proyecto permite monitorear IBM FlashSystem desde Zabbix mediante:

1. **Conexión SSH segura** al sistema de almacenamiento
2. **Ejecución de comandos** `svcinfo` de solo lectura
3. **Análisis automático** de la salida de los comandos
4. **Entrega de métricas** en formato compatible con Zabbix
5. **Descubrimiento dinámico** de componentes para reducir configuración manual

El sistema está diseñado para ser **estable en producción**, con manejo adecuado de errores, timeouts, reintentos y logging completo para troubleshooting.