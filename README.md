# MDML
Manufacturing Data &amp; Machine Learning Layer, Argonne National Laboratory


## Installation
Changes to the get_secrets.sh file will be required. Supply your own passwords in whichever manner you need. 
```
make build_docker_images
source get_secrets.sh
```
Make sure to supply a .s3cfg file in the mdml_register folder if connecting to the BIS S3 object store.

## Before Starting the MDML
Edit the nginx.conf file in the nginx folder. Name of the host must be changed.

## Starting the MDML
```
docker-compose up
```
