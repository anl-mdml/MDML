# MDML
Manufacturing Data &amp; Machine Learning Layer, Argonne National Laboratory


## Installation
Changes to the get_secrets.sh file will be required. Supply your own passwords in whichever manner you need. 
```
make docker_build_images
source get_secrets.sh
```
Make sure to supply a .s3cfg file in the mdml_register folder if connecting to the BIS S3 object store.

## Before Starting the MDML
Edit the grafana.ini file in the grafana folder. Specifically the fields: domain, root_url, cert_file, cert_key, database:url.
Edit the nginx.conf file in the nginx folder. Name of the host must be changed. 

## Starting the MDML
Changes to the docker-compose.yaml file may be required if your SSL certs are in a different location.
```
docker-compose up
```
