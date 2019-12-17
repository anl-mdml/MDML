# MDML
Manufacturing Data &amp; Machine Learning Layer, Argonne National Laboratory

## Installation

Changes to the get_secrets.sh file will be required. Supply your own passwords in whichever manner you need. 

```
make docker_build_images
source get_secrets.sh
docker-compose up
```
