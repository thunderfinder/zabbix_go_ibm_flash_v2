# Añadir permisos de ejecución al script
chmod +x scripts/flashsystem_collector.sh

# Mover el script al directorio de scripts externos de Zabbix
sudo cp scripts/flashsystem_collector.sh /usr/lib/zabbix/externalscripts/

# Asegurar que el script pertenezca al usuario zabbix
sudo chown zabbix:zabbix /usr/lib/zabbix/externalscripts/flashsystem_collector.sh

# Crear directorio de logs si no existe
sudo mkdir -p /var/log/zabbix

# Asegurar permisos adecuados para el directorio de logs
sudo chown zabbix:zabbix /var/log/zabbix
sudo chmod 755 /var/log/zabbix