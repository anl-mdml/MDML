# Manufacturing Data & Machine Learning Layer (MDML)


The Manufacturing Data & Machine Learning Layer was built to support scientists and their research efforts in the Materials Engineering Research Facility (MERF). MDML connects lab stations with HPC resources, AWS cloud, and local nodes for edge computing. Etc...


## Software Stack

* ### MQTT
    MDML uses the MQTT protocol for passing data between machines. A server located in MERF is running an Eclipse Mosquitto message broker and is responsible for this data sending. MQTT uses a publisher/subscriber model. To explain, a subscriber connects to the message broker and provides a topic string for receiving messages. When a publisher connects and publishes data, any client subscribed to the same topic will receive the message. Each message sent through MQTT consists of two parts supplied by the publisher: a *topic* and a *payload* containing the actual data. Topic strings can be hierarchical in nature, for example **MERF/FlameSpray/Device1**. MERF and FlameSpray are the first and second levels in this topic's hierarchy. Wildcards (#) can be included to receive all messages. For example, a subscriber with topic **MERF/FlameSpray/#** would receive all messages in which the topic starts with **MERF/FlameSpray/**.

* ### Node-RED
    Node-RED is a flow-oriented programming tool. Flows are editted in a web browser and deployed in place. Node-RED supports data connections to MQTT brokers. Within the MDML flow, messages are routed based on their topic strings. Actions can be applied to the data via metadata supplied in the payload. Functions are created in javascript to parse, analyze and/or store the data. Node-RED was chosen to provide a visual representation of the MDL framework and for its quick but effective development.

* ### InfluxDB
    InfluxDB is a time series database. Measurements are created for various devices and other data producers. InfluxDB stores a timestamp with each new value and keeps a record of old values up to a configurable point. The syntax of querying InfluDB is not exactly SQL but is extremely similar.

* ### Grafana
    Grafana enables easy dashboard creation by supplying prebuilt graphs and widgets. The only requirement to start creating dashboards is a connection to a supported data source. Supported data sources include: InfluxDB, MySQL, ElasticSearch, PostgreSQL, AWS CLoudWatch, Azure Monitor, and more. Grafana was chosen to give scientists the fexibility to tailor dashboards to their needs and provide near real-time monitoring of their experiments.

## Using MDML


* ### Connecting to MDML's message broker
    A connection to the broker can be made from any language with an MQTT client library. In python, the paho-mqtt library makes this connection. [TODO CHANGE AFTER SERVER INSTALL] The host IP address is 146.137.10.50 at port 1883.


* ### MDML's MQTT Message Format

    * #### Topic
            MERF/[Experiment]/[Action]/[Device]
        The structure of your topic should follow the format above where [Action], [Experiment], and [Device] are replaced with real values. __This format is essential__ so that collisions are avoided in the topic space and messages are properly routed. You __MUST__ be consistent with your experiment and device strings. Topic strings are __case-sensitive__.
        * `Actions`: 
          * CONFIG - Initializing experiment variables (data headers)
          * DATA - Used when sending data
          * RESET - Reset the system's state for a new experiment

    * #### Payloads
        The payload of each message must be a string. Stringifying dictionaries/objects to use as message payloads is perfectly acceptable - and also required. MDL requires this so information about the actions to be taken are coupled with the data itself. Below is the syntax of the message payloads respective to the `action` used in the topic string.

        * ##### CONFIG
            When sending a message with the action `CONFIG`, the device value in the topic string can be omitted because the configuration is stored at the experiment level. This configuration will be saved as metadata along with any data generated during the experiment. It is in your best interest to carefully create a configuration so that any necessary metadata will be included. The amount of detail included here will contribute to the usefulness and longevity of the dataset. Imagine you are a researcher trying to fully understand the datasets produced by this experiment. Below is an example configuration for an experiment using 2 devices. A breakdown of each field follows the example.

                {
                    'experiment': {
                        'experiment_id': 'FLAME',
                        'experiment_number': '1',
                        'experiment_notes': 'First run',
                        'experiment_devices': ['DEVICE_A', 'DEVICE_B']
                    },
                    'devices': [
                        {
                            'device_id': 'DEVICE_A',
                            'device_name': 'ANDOR Kymera328',
                            'device_version': '1.2',
                            'device_output': '2048 intensity values in the 250-700nm wavelength range',
                            'device_output_rate': 10,
                            'device_notes': 'Points directly at the flame in 8 different locations',
                            'headers': ["header 1", "header 2", ...],
                            'data_types': ["string", "float", ...],
                            'data_units': ["nanometer", "text", ...],
                            'save_tsv': true
                        },
                        ...
                        {
                            'device_id': 'DEVICE_Z'
                            'device_name': 'Scanning Mobility Particle Sizer',
                            'device_version': '5.3',
                            'device_output': 'Particle diameter size distribution',
                            'device_output_rate': 0.1,
                            'device_notes': 'Particles are split off from the exhaust of the furnace',
                            'headers': ["header 1", "header 2", ...],
                            'data_types': ["float", "string", ...],
                            'data_units': ["text", "count", ...],
                            'save_tsv': true
                        }
                    ]
                }
            * `experiment` - List of details about the experiment
              * `experiment_id` - Identifier for this experiment - used internally by the MDML
              * `experiment_number` - Number of this experiment in a possible series of experiments
              * `experiment_notes` - Any miscellaneous notes regarding the experiment
              * `experiment_devices` - List of device IDs used in the experiment
            * `devices` - List of devices that are sending data to MDML
              * `device_id` - Identifier for the device - used internally by the MDML
              * `device_name` - Technical name of the device/sensor that is outputting data
              * `device_version` - Current version of the device
              * `device_output` - Description of the output data
              * `device_output_rate` - Rate at which data is output (in hertz)
              * `device_notes` - Any extra notes that the user would like to include
              * `headers` - List of data headers
              * `data_types` - List of the data type to be sent
              * `data_units` - Units associated with the data (seconds, nanometers, text, count, etc.)
              * `save_tsv` - Boolean value (true/false) to save the data to a tab separated values file (tarball of all device data files will be output upon RESET command)

        * ##### DATA
                {
                    'data': [Actual data string],
                    'data_delimiter': [Delimiter used in the data string],
                    'data_type': [see data_type options]
                    'influx_measurement': [InfluxDB measurement name],
                }
            * `data` - The actual data generated to be stored, analyzed, etc.
            * `data_delimiter` - (optional) Describes how internal MDL functions will split the message's `data` and `data_headers` strings. If omitted, the system assumes the data string is only one value and is stored as such.
            * `data_type` - (optional) one of these values: "text/numeric", "image" (defaults to "text/numeric")
            * `influx_measurement` - (optional) A string to create a measurement in InfluxDB that the data will be stored under. This string will automatically be prefixed with the experiment value used in the topic string. If omitted, the data will not be stored in InfluxDB.

        * ##### RESET
            Sending this reset message will archive any files output during the experiment. It also performs some behind the scenes housekeeping to make sure all data message have made it fully through the pipeline.
                {
                    'reset': 1
                }
  
* ### Starting an experiment
    In order to begin a new experiment to start recording data, creating dashboards, and running analyses, a specific MQTT message must be sent to the MDL system. **TODO** finish this section with an example message after implementation.

* ### Stopping an experiment
  It is important to explicitly end an experiment so that data from multiple runs are not stored in one file.

* ### Receiving updates during the experiment relating to what MDML is doing.
  Any updates that the MDML is capable of providing to the researcher will be published to the same broker that data is sent through. To receive these updates, create an MQTT subscriber client that listens to 146.137.10.50 on port 1883 for the topic 'UPDATE/[experiment_id]'.