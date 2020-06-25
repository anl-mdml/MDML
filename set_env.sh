#!/bin/bash

# Change these lines as necessary
export MDML_HOST="merfpoc.egs.anl.gov"
export PRIVATE_KEY_PATH="/etc/ssl/private/nginx-selfsigned.key"
export CERT_PATH="/etc/ssl/certs/nginx-selfsigned.crt"

# Allow access to the TEST experiment by default.  
export ALLOW_TEST=false # or true

# Allow automatic creation of experiment IDs
export AUTO_CREATE_IDS=false # or true


# All admin passwords are retrieved from AWS' Secrets Manager.
# Change the passwords here as necessary
SECRETS=$(aws2 secretsmanager get-secret-value --secret-id MDML/merfpoc | jq --raw-output '.SecretString')
export MDML_INFLUXDB_SECRET=$(echo $SECRETS | jq -r '.influxdb_secret')
export MDML_GRAFANA_SECRET=$(echo $SECRETS | jq -r '.grafana_secret')
export MDML_MINIO_SECRET=$(echo $SECRETS | jq -r '.minio_secret')
export MDML_GRAFDB_SECRET=$(echo $SECRETS | jq -r '.grafdb_secret')
export MDML_GRAFDB_ROOT_SECRET=$(echo $SECRETS | jq -r '.grafdb_root_secret')
export MDML_NODE_RED_PASS=$(echo $SECRETS | jq -r '.node_red_admin')
export MDML_NODE_MQTT_USER=$(echo $SECRETS | jq -r '.node_red_mqtt_user')

# Create credentials config file for the Minio object store
python ./mdml_register/create_minio_config.py $MDML_MINIO_SECRET

# Create credentials file for Node-RED Admin - Requires npm to be installed with either the bcryptjs module 
export NODE_PATH=/usr/lib/node_modules # needed as bcryptjs module was not being found
node -e "console.log(require('bcryptjs').hashSync(process.argv[1], 8));" $MDML_NODE_RED_PASS | tr -d '\n' > ./node_red/data/node_red_admin_creds.txt
