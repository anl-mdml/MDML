# MDML
Manufacturing Data &amp; Machine Learning Layer, Argonne National Laboratory


## Installation
The MDML runs with a docker container for each component. All of the docker containers can be build with ```make build_docker_images```. Before starting the containers, certain environment variables must be created. Editting the ```set_env.sh``` file will be required to properly set admin passwords, key file locations, and more. The MDML uses AWS' Secrets Manager to import passwords so they are not hard coded. Passwords can still be hard coded into the set_env.sh file, but AWS compenents and data parsing should be replaced. Once the set_env.sh has been changed, run ```source set_env.sh```.

## Before Starting the MDML
Edit the nginx.conf file in the nginx folder. Host names will need to be changed throughout.

## Starting the MDML
```
docker-compose up
```
