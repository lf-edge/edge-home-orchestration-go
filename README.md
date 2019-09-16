Edge Orchestration
=================================

[![Goreport](https://goreportcard.com/badge/github.com/lf-edge/edge-home-orchestration-go)](https://goreportcard.com/report/github.com/lf-edge/edge-home-orchestration-go)

The main purpose of Edge Orchestration is to implement distributed computing between Docker Container enabled devices. 

The Workflow is as follows:

0. Run Edge Orchestration container at a host machine. Then the host machine becomes an Edge Orchestration device. 
1. The device receives Service Execution request from host via REST API.
2. The device gets and compares the **scores** from the other Edge Orchestration Devices.
3. The device requests Service Execution to the device that has the highest score.
   - If the device itself has the highest score, the service is executed on itself.   

## Quick start ##
This section provides how to download and run pre-built Docker image without building the project.

#### 1. Install docker-ce ####
- docker-ce
  - Version: 17.09 (or above)
  - [How to install](https://docs.docker.com/engine/installation/linux/docker-ce/ubuntu/)

#### 2. Download Docker image ####
Please download [edge-orchestration](https://github.com/lf-edge/edge-home-orchestration-go/releases/download/alpha/edge-orchestration.tar)

#### 3. Load Docker image from tar file ####
```shell
$ docker load -i edge-orchestration.tar
```
If it succeeds, you can see the Docker image as follows:
```shell
$ docker images
REPOSITORY                 TAG                 IMAGE ID            CREATED             SIZE
edge-orchestration         alpha               533ae9f5f9a5        3 minutes ago       131MB
```

#### 4. Add User ID #### 

To let the Edge Orchestration devices communicate with each other, each devices should have same authentication key in:

/etc/edge-orchestration/orchestration_userID.txt
```shell
$ cat /etc/edge-orchestration/orchestration_userID.txt

thisIsAuthenticationKey
```
*Any string can be authentication key

#### 5. Run with Docker image ####
You can execute Edge Orchestration with a Docker image as follows:

```shell
$ docker run -it -d \
                --privileged \
                --network="host" \
                --name edge-orchestration \
                -v /var/run/:/var/run/:rw \
                -v /var/log/:/var/log/:rw \
                -v /etc/:/etc/:rw \
                -v /usr/bin/docker:/usr/bin/docker \
                edge-orchestration:alpha
``` 

## Build Prerequisites ##
- docker-ce
    - Version: 17.06 (or above)
    - [How to install](https://docs.docker.com/engine/installation/linux/docker-ce/ubuntu/)
- go compiler
    - Version: 1.8 (or above)
    - [How to install](https://golang.org/dl/)

## How to build ##

This project offers Docker image build as a default build option. 

```shell
$ ./build.sh 
```
This build script will build Edge Orchestration in Docker Container environment. 

If it succeeds, you can see the container runs as follows:
```shell
**********************************
 Run Docker container 
**********************************
5d2efd81057fe56236602acfece0e8f11d447b54627f4f3669b18c85a95b8687
CONTAINER ID        IMAGE                      COMMAND             CREATED                  STATUS                  PORTS               NAMES
5d2efd81057f        edge-orchestration:alpha   "sh run.sh"         Less than a second ago   Up Less than a second                       edge-orchestration
```

and the built image as follows:
```shell
$ docker images
REPOSITORY                 TAG                 IMAGE ID            CREATED             SIZE
edge-orchestration         alpha               533ae9f5f9a5        3 seconds ago       131MB
```

- All Build Options

```shell
build script
Usage:
--------------------------------------------------------------------------------------
  ./build.sh                  : build edge-orchestration by default container
  ./build.sh container        : build Docker container as build system environmet
  ./build.sh object           : build edge-orchestration archive for c and java
  ./build.sh clean            : build clean
  ./build.sh test [PKG_NAME]  : run unittests (optional for PKG_NAME)
--------------------------------------------------------------------------------------
```
 
## API Document ##
Edge Orchestration provides only one REST API for its operation. Description for the APIs are stored in <root>/doc folder.
- **[edge_orchestration_api.yaml](./doc/edge_orchestration_api.yaml)**

Note that you can visit [Swagger Editor](https://editor.swagger.io/) to graphically investigate the REST API in YAML.

## How to work ##
#### 0. Prerequisites ####
  - 2 or more devices with Ubuntu 14.04 (or above) and Docker 17.09 (or above)
  - Same WIFI network connected among the devices.
  - Same Authentication key in /etc/edge-orchestration/orchestration_userID.txt
    - Please see the above [4. Add User ID](#4-add-user-id) to know how to add authentication key
  - Edge Orchestration Docker image
    - Please see the above [How to build](#how-to-build) to know how to build Edge Orchestration Docker image


#### 1. Run Edge Orchestration container #### 

```shell
$ docker run -it -d \
                --privileged \
                --network="host" \
                --name edge-orchestration \
                -v /var/run/:/var/run/:rw \
                -v /var/log/:/var/log/:rw \
                -v /etc/:/etc/:rw \
                -v /usr/bin/docker:/usr/bin/docker \
                edge-orchestration:alpha
``` 
- Result 

```shell
$ docker logs -f edge-orchestration 

2019/06/07 05:45:29 main.go:102: [interface] OrchestrationInit
2019/06/07 05:45:29 orchestration.go:163: Orchestration Start In
2019/06/07 05:45:29 networkhelper.go:152: >>  192.168.1.37
2019/06/07 05:45:29 networkhelper.go:152: >>  fe80::f112:19a8:eca4:724f
2019/06/07 05:45:29 discovery.go:317: [discoverymgr] Platform:: docker  OnboardType:: container
2019/06/07 05:45:29 discovery.go:312: [discoverymgr] UUID :  80924bc5-afc8-4135-85c2-105bf9a5a5c4
2019/06/07 05:45:29 discovery.go:370: [discoverymgr] ip : 192.168.1.37
2019/06/07 05:45:29 discovery.go:455: [discoverymgr] [mdnsTXTSizeChecker] size ::  32  Bytes
2019/06/07 05:45:29 orchestration.go:170: Orchestration Start Out
2019/06/07 05:45:29 route.go:71: {APIV1ServicemgrServicesPost POST /api/v1/servicemgr/services 0x7d35c0}
2019/06/07 05:45:29 route.go:71: {APIV1ServicemgrServicesNotificationServiceIDPost POST /api/v1/servicemgr/services/notification/{serviceid} 0x7d3620}
2019/06/07 05:45:29 route.go:71: {APIV1ScoringmgrScoreLibnameGet GET /api/v1/scoringmgr/score 0x7d3680}
2019/06/07 05:45:29 route.go:71: {APIV1RequestServicePost POST /api/v1/orchestration/services 0x7d1e30}
2019/06/07 05:45:29 main.go:151: interface orchestration init done
2019/06/07 05:45:29 route.go:63: ListenAndServe
```

#### 2. Request to execute a service ####

RESTAPI
- POST  
- **IP:56001/api/v1/orchestration/services** 
- BODY : 
    ```json
    {
        "ServiceName": "hello-world",
        "ServiceInfo": [
        {
            "ExecutionType": "container",
            "ExecCmd": [
                "docker",
                "run",
                "-v", "/var/run:/var/run:rw",
                "hello-world"
            ]
        }],
        "StatusCallbackURI": "http://localhost:8888/api/v1/services/notification"
    }
    ```
- Curl Example:
    ```json
  - curl -X POST "IP:56001/api/v1/orchestration/services" -H "accept: application/json" -H "Content-Type: application/json" -d "{ \"ServiceName\": \"hello-world\", \"ServiceInfo\": [{ \"ExecutionType\": \"container\", \"ExecCmd\": [ \"docker\", \"run\", \"-v\", \"/var/run:/var/run:rw\", \"hello-world\"]}], \"StatusCallbackURI\": \"http://localhost:8888/api/v1/services/notification\"}"
    ```

- Result(Execution on itself)
  
  ```shell
  $ docker logs -f edge-orchestration 

  2019/06/07 05:41:03 externalhandler.go:75: [RestExternalInterface] APIV1RequestServicePost
  2019/06/07 05:41:03 orchestration_api.go:70: [RequestService] container_service: [docker run -v /var/run:/var/run:rw hello-world]
  2019/06/07 05:41:03 scoringmgr.go:131: [IN] getScoreLocalEnv
  2019/06/07 05:41:03 scoringmgr.go:139: scoringmgr scoreValue :  7.481732534124991
  2019/06/07 05:41:03 orchestration_api.go:90: [orchestrationapi]  [{192.168.1.37 7.481732534124991}]
  2019/06/07 05:41:03 route.go:87: POST /api/v1/orchestration/services APIV1RequestServicePost 272.182µs
  2019/06/07 05:41:03 containerexecutor.go:75: [containerexecutor] container_service [docker run -v /var/run:/var/run:rw hello-world]
  2019/06/07 05:41:03 containerexecutor.go:76: [containerexecutor] parameter length : 5
  2019/06/07 06:46:12 route.go:87: POST /api/v1/orchestration/services APIV1RequestServicePost 396.063µs
  {"status":"Pulling from library/hello-world","id":"latest"}
  {"status":"Digest: sha256:0e11c388b664df8a27a901dce21eb89f11d8292f7fca1b3e3c4321bf7897bffe"}
  {"status":"Status: Image is up to date for hello-world:latest"}
  2019/06/07 05:41:21 containerexecutor.go:90: [containerexecutor] create container : bb8c3425ec
  2019/06/07 05:41:22 containerexecutor.go:108: [containerexecutor] container execution status : 0
  2019-06-07T05:41:22.144893108Z 
  2019-06-07T05:45:35.717107646Z Hello from Docker!
  2019-06-07T05:45:35.717110637Z This message shows that your installation appears to be working correctly.
  2019-06-07T05:45:35.717112612Z 
  2019-06-07T05:45:35.717114418Z To generate this message, Docker took the following steps:
  2019-06-07T05:45:35.717116284Z  1. The Docker client contacted the Docker daemon.
  2019-06-07T05:45:35.717118076Z  2. The Docker daemon pulled the "hello-world" image from the Docker Hub.
  2019-06-07T05:45:35.717120060Z     (amd64)
  2019-06-07T05:45:35.717121906Z  3. The Docker daemon created a new container from that image which runs the
  2019-06-07T05:45:35.717123788Z     executable that produces the output you are currently reading.
  2019-06-07T05:45:35.717125570Z  4. The Docker daemon streamed that output to the Docker client, which sent it
  2019-06-07T05:45:35.717127407Z     to your terminal.
  2019-06-07T05:45:35.717129190Z 
  2019-06-07T05:45:35.717130971Z To try something more ambitious, you can run an Ubuntu container with:
  2019-06-07T05:45:35.717132780Z  $ docker run -it ubuntu bash
  2019-06-07T05:45:35.717134548Z 
  2019-06-07T05:45:35.717136249Z Share images, automate workflows, and more with a free Docker ID:
  2019-06-07T05:45:35.717138053Z  https://hub.docker.com/
  2019-06-07T05:45:35.717139826Z 
  2019-06-07T05:45:35.717141538Z For more examples and ideas, visit:
  2019-06-07T05:45:35.717143307Z  https://docs.docker.com/get-started/
  2019-06-07T05:45:35.717145081Z 
  2019/06/07 05:41:22 orchestration_api.go:163: [orchestrationapi] service status changed [appNames:container_service][status:Finished]
  ```
- Not supported docker run option [*Args* in Body]
   ```--detach, -d
   --detach-keys
   --disable-content-trust
   --sig-proxy
   --name
   --platform
   --help
   --cpu-percent       (Windows only option)
   --cpu-count         (Windows only option)
   --io-maxbandwidth   (Windows only option)
   --io-maxiops        (Windows only option)
   ```

