# MDML
Manufacturing Data &amp; Machine Learning Layer, Argonne National Laboratory


## Installation
Changes to the get_secrets.sh file will be required. Supply your own passwords in whichever manner you need. 
```
make docker_build_images
source get_secrets.sh
```
Make sure to supply a .s3cfg file in the mdml_register folder if connecting to the BIS S3 object store.


## Starting the MDML
Changes to the docker-compose.yaml file may be required if your SSl certs are in a different location.
```
docker-compose up
```
