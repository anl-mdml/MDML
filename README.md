# MDML
Manufacturing Data &amp; Machine Learning Layer, Argonne National Laboratory


## Installation
```
make build_docker_images
```
The MDML runs with a docker container for each component/esrvice. All of the docker containers can be built with the command above. Before starting the containers, certain environment variables must be created. Editting the ```set_env.sh``` file will be required to properly set admin passwords, key file locations, and more. The MDML in use at ANL uses AWS' Secrets Manager to import passwords so that they are not hard coded. Passwords can be hard coded into the set_env.sh file, but AWS components and data parsing should be replaced. Once the set_env.sh has been changed for your setup, run:
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
Once the MDML is started a user account must be created in order to use the MDML. This can be done through the home page (https://your_host_name). When a user account is created, an account is created on the Mosquitto (MQTT) broker, Grafana instance, and MinIO object store. Due to the authentication flow of Mosquitto, the broker must be restarted to start accepting new user accounts attempting to send messages. This can be done by simply running ```docker-compose down``` followed by ```docker-compose up```. This requirement is something that we are working to remove from the MDML as it interrupts other users streaming data.

Another manual step an admin of the MDML will need to take before running an experiment is adding an experiment ID to the allowable experiment IDs in NodeRED. To access NodeRED, navigate to http://your_host_name:1880/admin. After logging in with the password supplied in the set_env.sh file, look for the node titled "Add new experiment ID with payload". Double click the title to open the node's settings. In the payload box, enter the name of the experiment that you want to add. Click done to close the node's setting, and then click deploy in the upper right corner. The final step is to click the button on the left side of the "Add new experiment ID with payload" node. This injects a message into the flow with the payload you have set and adds it to the allowable experiment IDs.

At this point, the new user will be able to log in and start an experiment with the username, password, and experiment ID they used during registration. By default, a new user will only be granted access to their given experiment ID as well as the TEST experiment ID for running examples.

