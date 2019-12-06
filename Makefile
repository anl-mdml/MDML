build_docker_images:
	docker build -t grafana grafana/.
	docker build -t grafana_mysqldb grafana_mysqldb/.
	docker build -t influxdb influxdb/.
	docker build -t mdml_register mdml_register/.
	docker build -t minio minio/.
	docker build -t mosquitto mosquitto/.
	docker build -t nginx nginx/.
	docker build -t node_red node_red/.
