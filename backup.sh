DATE=$(date +"%m_%d_%Y")

# Create folder for this backup
mkdir backups/${DATE}

# Backup Grafana Dashboards and grab the SQL file.
docker exec -it mdml_grafana_mysqldb_1 bash -c "mysqldump -u grafana -p grafana > /var/lib/backups/merf_mdml_graf_backup_${DATE}.sql"
cp grafana_mysqldb/backups/merf_mdml_graf_backup_${DATE}.sql backups/${DATE}

# Grab MQTT message broker user account and ACL 
cp mosquitto/acl_file.txt backups/${DATE}
cp mosquitto/wordpassfile.txt backups/${DATE}

# Grab object store data (.minio.sys holds user info)
cp -r minio/mnt/data backups/${DATE}

# Grab Nginx config file
cp nginx/nginx.conf backups/${DATE}