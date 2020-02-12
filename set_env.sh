#!/bin/bash

# Change these lines as necessary
export MDML_HOST="146.139.77.100"
export PRIVATE_KEY_PATH="/etc/pki/nginx/private/server.key"
export CERT_PATH="/etc/pki/nginx/server.crt"

# All admin passwords are retrieved from AWS' Secrets Manager.
# Change the passwords here as necessary
SECRETS=$(aws secretsmanager get-secret-value --secret-id MDML/merfpoc | jq --raw-output '.SecretString')
export MDML_INFLUXDB_SECRET=$(echo $SECRETS | jq -r '.influxdb_secret')
export MDML_GRAFANA_SECRET=$(echo $SECRETS | jq -r '.grafana_secret')
export MDML_MINIO_SECRET=$(echo $SECRETS | jq -r '.minio_secret')
export MDML_GRAFDB_SECRET=$(echo $SECRETS | jq -r '.grafdb_secret')
export MDML_GRAFDB_ROOT_SECRET=$(echo $SECRETS | jq -r '.grafdb_root_secret')

# Creates credentials config file for the Minio object store
python3 ./mdml_register/create_minio_config.py $MDML_MINIO_SECRET