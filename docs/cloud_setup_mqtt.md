# Cloud Synchronization
## Contents
1. [Introduction](#1-introduction)
2. [Local Setup](#2-Local-Setup-Description)
3. [Getting Started With CloudMQTT](#3-Getting-Started-With-CloudMQTT)
4. [CloudMQTT Setup on Heroku](#4-CloudMQTT-Setup-on-Heroku)
5. [Optional - Testing Setup on A Public Broker - HiveMQ](#5-how-to-setup)
6. [Links for setup on other vendors](#6-Links-for-setup-on-other-vendors)

## 1. Introduction
This module would be responsible for sending the data (can be sensor data or any reading or image/video data, etc.) collected from different devices in the home environment to a cloud Endpoint. This Sync to the cloud will be in every few seconds. The user can define the endpoints and provide the configuration for automatic sync to the cloud.

## 2. Local Setup Description
Data export to an external source, like an MQTT topic, AWS or Azure, can be done by configuring the options in the
[docker-compose.yml](https://github.com/lf-edge/edge-home-orchestration-go/blob/da65e9430e44a678e6a8b00d9bb7b5c5a51f066c/deployments/datastorage/docker-compose.yml) in deployments/datastorage.

### Stop the running containers (If they are running) to modify the configuration
```
docker-compose stop 
```
### Open the file in your favourite editor
```
sudo gedit docker-compose.yml
```
### Uncomment the app-service-mqtt part (Example is given at the end)

#### Required Parameters
```
WRITABLE_PIPELINE_FUNCTIONS_MQTTSEND_ADDRESSABLE_ADDRESS: [Your Address]
```
#### Optional Parameters
```
environment:
  - WRITABLE_PIPELINE_FUNCTIONS_MQTTSEND_ADDRESSABLE_PORT: ["your port"]
  - WRITABLE_PIPELINE_FUNCTIONS_MQTTSEND_ADDRESSABLE_PROTOCOL: [tcp or tcps]
  - WRITABLE_PIPELINE_FUNCTIONS_MQTTSEND_ADDRESSABLE_PUBLISHER: [your name]
  - WRITABLE_PIPELINE_FUNCTIONS_MQTTSEND_ADDRESSABLE_USER: [your username]
  - WRITABLE_PIPELINE_FUNCTIONS_MQTTSEND_ADDRESSABLE_PASSWORD: [your password]
  - WRITABLE_PIPELINE_FUNCTIONS_MQTTSEND_ADDRESSABLE_TOPIC: [your topic]
  - WRITABLE_PIPELINE_FUNCTIONS_MQTTSEND_PARAMETERS_QOS: ["your quality or service"]
  - WRITABLE_PIPELINE_FUNCTIONS_MQTTSEND_PARAMETERS_KEY: [your Key]  
  - WRITABLE_PIPELINE_FUNCTIONS_MQTTSEND_PARAMETERS_CERT: [your Certificate]
  - WRITABLE_PIPELINE_FUNCTIONS_MQTTSEND_PARAMETERS_AUTORECONNECT: ["true" or "false"]
  - WRITABLE_PIPELINE_FUNCTIONS_MQTTSEND_PARAMETERS_RETAIN: ["true" or "false"]
  - WRITABLE_PIPELINE_FUNCTIONS_MQTTSEND_PARAMETERS_PERSISTONERROR: ["true" or "false"]

```
#### Example Configuration used in testing
```
app-service-mqtt:
    image: edgexfoundry/docker-app-service-configurable:1.3.1
    ports:
      - "0.0.0.0:48101:48101"
    container_name: edgex-app-service-configurable-mqtt
    hostname: edgex-app-service-configurable-mqtt
    networks:
      edgex-network:
        aliases:
          - edgex-app-service-configurable-mqtt
    environment:
      EDGEX_SECURITY_SECRET_STORE: "false"
      REGISTRY_HOST: edgex-core-consul
      CLIENTS_COREDATA_HOST: edgex-core-data
      CLIENTS_DATA_HOST: edgex-core-data # for device services
      CLIENTS_NOTIFICATIONS_HOST: edgex-support-notifications
      CLIENTS_METADATA_HOST: edgex-core-metadata
      CLIENTS_COMMAND_HOST: edgex-core-command
      CLIENTS_SCHEDULER_HOST: edgex-support-scheduler
      CLIENTS_RULESENGINE_HOST: edgex-kuiper
      DATABASES_PRIMARY_HOST: edgex-redis
      EDGEX_PROFILE: mqtt-export
      SERVICE_HOST: edgex-app-service-configurable-mqtt
      SERVICE_PORT: 48101
      MESSAGEBUS_SUBSCRIBEHOST_HOST: edgex-core-data
      BINDING_PUBLISHTOPIC: events
      # Added for MQTT export using app service
      WRITABLE_PIPELINE_FUNCTIONS_MQTTSEND_ADDRESSABLE_ADDRESS: broker.mqttdashboard.com
      WRITABLE_PIPELINE_FUNCTIONS_MQTTSEND_ADDRESSABLE_PORT: 1883
      WRITABLE_PIPELINE_FUNCTIONS_MQTTSEND_ADDRESSABLE_PROTOCOL: tcp
      WRITABLE_PIPELINE_FUNCTIONS_MQTTSEND_ADDRESSABLE_TOPIC: "DataStorage"
      WRITABLE_PIPELINE_FUNCTIONS_MQTTSEND_PARAMETERS_AUTORECONNECT: "true"
      WRITABLE_PIPELINE_FUNCTIONS_MQTTSEND_PARAMETERS_RETAIN: "true"
      WRITABLE_PIPELINE_FUNCTIONS_MQTTSEND_PARAMETERS_PERSISTONERROR: "false"
      # WRITABLE_PIPELINE_FUNCTIONS_MQTTSEND_ADDRESSABLE_PUBLISHER:
      # WRITABLE_PIPELINE_FUNCTIONS_MQTTSEND_ADDRESSABLE_USER:
      # WRITABLE_PIPELINE_FUNCTIONS_MQTTSEND_ADDRESSABLE_PASSWORD:
      # WRITABLE_PIPELINE_FUNCTIONS_MQTTSEND_PARAMETERS_QOS: ["your quality or service"]
      # WRITABLE_PIPELINE_FUNCTIONS_MQTTSEND_PARAMETERS_KEY: [your Key]
      # WRITABLE_PIPELINE_FUNCTIONS_MQTTSEND_PARAMETERS_CERT: [your Certificate]
    depends_on:
      - consul
      # - logging  # uncomment if re-enabled remote logging
      - data
```

For more details please refer to the [Official Guide](https://docs.edgexfoundry.org/1.3/getting-started/quick-start/#exporting-data).

## 3. Getting Started With CloudMQTT

CloudMQTT are managed Mosquitto servers in the cloud. Mosquitto implements the MQ Telemetry Transport protocol, MQTT, which provides lightweight methods of carrying out messaging using a publish/subscribe message queueing model.

![Producer Consumer](images/CloudSync/ProducerConsumer.png)

MQTT is the machine-to-machine protocol of the future. It is ideal for the “Internet of Things” world of connected devices. Its minimal design makes it perfect for built-in systems, mobile phones and other memory and bandwidth sensitive applications.

Message queues provide an asynchronous communications protocol, the sender and receiver of the message do not need to interact with the message queue at the same time. Messages placed onto the queue are stored until the recipient retrieves them or until the messages times out. MQTT and Mosquitto are for good use by bandwidth sensitive applications.

CloudMQTT let you focus on the application instead of spending time on scaling the broker or patching the platform.

![Role of MQTT](images/CloudSync/RoleOfMQTT.png)

## 4. CloudMQTT Setup on Heroku

#### For Basic Understanding and setup of Heroku CLoud please visit https://devcenter.heroku.com/

#### Setting Up CloudMQTT 
CloudMQTT is an add-on for providing a MQTT broker to your application(s).
MQTT is a lightweight pub/sub protocol, especially suited for low processor/bandwidth units like sensors and built-in system, but also suited for fast communication within applications.

#### Provisioning the add-on
```
heroku addons:create cloudmqtt
-----> Adding cloudmqtt to sharp-mountain-4005... done, v18 (free)
```

Once CloudMQTT has been added a `CLOUDMQTT_URL` setting will be available in the app configuration and will contain the canonical URL used to access the newly provisioned CloudMQTT service instance. This can be confirmed using the `heroku config:get` command.

```
heroku config:get CLOUDMQTT_URL
mqtt://user:pass@broker.cloudmqtt.com
```

#### Dashboard
The CloudMQTT dashboard allows you to monitor your current connection and message traffic statistics.

The dashboard can be accessed via the CLI:
```
heroku addons:open cloudmqtt
Opening cloudmqtt for sharp-mountain-4005...
```
or by visiting the [Heroku Dashboard](https://dashboard.heroku.com/) and selecting the application in question. Select CloudMQTT from the Add-ons menu.

# 5. Optional - Testing Setup on A Public Broker - HiveMQ

HiveMQ provides a web-based client that you can use to set up a free broker. Use a browser to go to the [client's URL](http://www.hivemq.com/demos/websocket-client/). Once there, hit the Connect button to connect to the HiveMQ public broker.

![Connect to HiveMQ](images/CloudSync/ConnectToHiveMQ.png)
Using the HiveMQ provided client tool, connect to the same public HiveMQ broker your configurable application service is sending data to.

Then, use the Subscriptions area to subscribe to your Topic Defined in the local configuration.

![HiveMQTT Messages](images/CloudSync/HiveMQTTWebClient.png)

You must subscribe to the same topic as in local configuration - to see the data sent by the configurable application service.

![Subscriptions](images/CloudSync/HiveMQTTMessages.png)

Once subscribed, the event data will begin to appear in the Messages area on the browser screen.

***Please Note :*** This is a public broker and should only be used for testing. Also the QOS for this would not be perfect so some messages might be lost in transition. This will not be a problem with paid cloud brokers.

# 6. Links for setup on other vendors

- [Azure](https://docs.microsoft.com/en-us/azure/iot-hub/iot-hub-mqtt-support)
- [AWS](https://docs.aws.amazon.com/iot/latest/developerguide/mqtt.html)