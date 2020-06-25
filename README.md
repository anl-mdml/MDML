# MDML
Manufacturing Data &amp; Machine Learning Layer, Argonne National Laboratory


## Installation
```
make build_docker_images
```
The MDML runs with a docker container for each component/service. All of the docker containers can be built with the command above. Before starting the containers, certain environment variables must be set. Editting the ```set_env.sh``` file will be required to properly set admin passwords, key file locations, and more. See the __Environment Variables__ section below for details. The MDML in use at Argonne National Laboratory uses AWS' Secrets Manager to import passwords so that they are not hard coded. Passwords can be hard coded into the set_env.sh file, but AWS components and data parsing should be removed. Once the set_env.sh has been changed for your setup, run:
```
source set_env.sh
```

## Before Starting the MDML
Edit the nginx.conf file in the nginx folder. Only host names will need to be changed throughout.

## Starting the MDML
```
docker-compose up
```
Docker compose is used to start all of the docker containers used by the MDML. The first time starting the MDML you may see errors that Grafana is exiting. This is expected as Grafana's backend MySQL database is still being initialized. Once the database is ready, Grafana should no longer quit.

## Common Problems During Setup
If nginx is repeatedly failing to start, the container may not be able to find the proper keys to enable HTTPS. Grafana may also fail here for the same reason.



## Administering the MDML
Once started, a user account must be created in order to use the MDML. This can be done through the home page (https://your_host_name). When a user registers, an account is create for them on the Mosquitto (MQTT) broker, Grafana instance, and MinIO object store. Due to the authentication flow of Mosquitto, the broker must be restarted to start accepting new user accounts attempting to send messages. This can be done by running ```docker-compose down``` (shuts down all docker containers) followed by ```docker-compose up``` (starts all docker containers). This requirement is something that we are working to remove from the MDML as it interrupts other users streaming data.

A new user will be able to log in and start an experiment with the username, password, and experiment ID they used during registration. By default, a new user will only be granted access to their given experiment ID as well as the TEST experiment ID for running examples.

## Environment Variables

#### ALLOW_TEST
A value of "true" (default) will allow new users to run experiments under the TEST experiment ID. A value of "false" will require the user to use the experiment ID they supplied during registration.

#### AUTO_CREATE_IDS
A value of "true" will tell the MDML to automatically allow new experiment IDs that are being registered. A value of "false" (default) will stop the MDML from automatically allowing new experiment IDs (more secure). In this case, an admin of the MDML will need to add an experiment ID to the allowable experiment IDs in NodeRED. To access NodeRED, navigate to http://your_host_name:1880/admin. After logging in with the password supplied in the set_env.sh file, look for the node titled "Add new experiment ID with payload". Double click to open the node's settings. In the payload text box, enter the name of the experiment that you want to add. Click Done to close the node's settings, and then click Deploy in the upper right corner. The final step is to click the button on the left side of the "Add new experiment ID with payload" node. This injects a message into the flow with the payload you have set and adds it to the allowable experiment IDs.

#### MDML_HOST
Host name of the server running the MDML instance. This is also the address you will connect to in your browser.

#### PRIVATE_KEY_PATH & CERT_PATH
Absolute file paths to the SSL certificate and private key. Instrusctions to create self-signed SSL certificates on [Ubuntu](https://www.digitalocean.com/community/tutorials/how-to-create-a-self-signed-ssl-certificate-for-nginx-in-ubuntu-16-04) and [CentOS 7](https://www.digitalocean.com/community/tutorials/how-to-create-a-self-signed-ssl-certificate-for-nginx-on-centos-7). Self-signed certificates will work fine, but users will most likely be prompted with a security warning by their browser.