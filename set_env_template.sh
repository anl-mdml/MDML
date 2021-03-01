
#!/bin/bash

# Allow access to the TEST experiment by default.  
export ALLOW_TEST= #false # or true

# Allow automatic creation of experiment IDs
export AUTO_CREATE_IDS= #false # or true

# Use BIS Object store
export USE_BIS= #false # or true

# CHANGE these lines as necessary
export MDML_HOST="[INSERT HOST NAME HERE]"
export PRIVATE_KEY_PATH="[PATH TO PRIVATE KEY]"
export CERT_PATH="[PATH TO CERT FILE]"



# Change the passwords here as necessary
# MDML uses AWS' Secrets Manager but you could hard code passwords here instead...
#SECRETS=$(aws secretsmanager get-secret-value --secret-id [SECRET ID] | jq --raw-output '.SecretString')
export MDML_INFLUXDB_SECRET=
export MDML_GRAFANA_SECRET=
export MDML_MINIO_SECRET=
export MDML_GRAFDB_SECRET=
export MDML_GRAFDB_ROOT_SECRET=
export MDML_NODE_RED_PASS=
export MDML_NODE_MQTT_USER=


### DO NOT CHANGE ANYTHING BELOW HERE ###
# Create credentials config file for the Minio object store
mkdir ./mdml_register/.mc
python ./mdml_register/create_minio_config.py $MDML_MINIO_SECRET

# Create credentials files for Node-RED to access MinIO
printf $MDML_MINIO_SECRET > ./node_red/data/minio_admin_creds.txt

# Create credentials file for Node-RED Admin - Requires npm to be installed with either the bcryptjs module 
export NODE_PATH=/usr/lib/node_modules # needed as bcryptjs module was not being found
node -e "console.log(require('bcryptjs').hashSync(process.argv[1], 8));" $MDML_NODE_RED_PASS | tr -d '\n' > ./node_red/data/node_red_admin_creds.txt
