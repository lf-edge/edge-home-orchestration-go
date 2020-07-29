# Edge Orchestration on x86_64 Linux

## Quick start
This section provides how to download and run pre-built Docker image without building the project.

#### 1. Install docker-ce
- docker-ce
  - Version: 17.09 (or above)
  - [How to install](https://docs.docker.com/engine/installation/linux/docker-ce/ubuntu/)

#### 2. Download Docker image
Please download [edge-orchestration docker container](https://github.com/lf-edge/edge-home-orchestration-go/releases/download/Baobab_rc1/edge-orchestration.tar)

#### 3. Load Docker image from tar file
```shell
$ docker load -i edge-orchestration.tar
```
If it succeeds, you can see the Docker image as follows:
```shell
$ docker images
REPOSITORY                 TAG                 IMAGE ID            CREATED             SIZE
edge-orchestration         baobab              502e3c07b01f        3 minutes ago       132MB
```

#### 4. Add Key file

To let the Edge Orchestration devices communicate with each other, each devices should have same authentication key in:
`/var/edge-orchestration/data/cert/edge-orchestration.key`

```shell
$ sudo cp {SampleKey.key} /var/edge-orchestration/data/cert/edge-orchestration.key
```
> Any cert file can be authentication key

#### 5. Run with Docker image
You can execute Edge Orchestration with a Docker image as follows:

```shell
$ docker run -it -d \
                --privileged \
                --network="host" \
                --name edge-orchestration \
                -v /var/edge-orchestration/:/var/edge-orchestration/:rw \
                -v /var/run/docker.sock:/var/run/docker.sock:rw \
                -v /proc/:/process/:ro \
                edge-orchestration:baobab
```
---

## How to build

#### Build Prerequisites
- docker-ce
    - Version: 17.06 (or above)
    - [How to install](https://docs.docker.com/engine/installation/linux/docker-ce/ubuntu/)

> For [execution of docker commands with non-root privileges](https://docs.docker.com/install/linux/linux-postinstall/#manage-docker-as-a-non-root-user) you need to add `$USER` to docker group.  
`$ sudo usermod -aG docker $USER`

- go compiler
    - Version: 1.10 (or above)
    - [How to install](https://golang.org/dl/)

> To build Edge Orchestrator from Go sources, you need to set GOPATH environment variable:  
`$ export GOPATH=$HOME/go`  
`$ export PATH=$PATH:$GOPATH/bin`

- glide
To install `glide`, you need to add appropriate repository to apt list. To do that, execute following shell commands:

```shell
$ sudo add-apt-repository -y ppa:masterminds/glide && sudo apt-get update
$ sudo apt-get install -y glide
```

- extra linux utilities
```
$ sudo apt-get install tree jq
```

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
5d2efd81057f        edge-orchestration:baobab  "sh run.sh"         Less than a second ago   Up Less than a second                       edge-orchestration
```

and the built image as follows:
```shell
$ docker images
REPOSITORY                 TAG                 IMAGE ID            CREATED             SIZE
edge-orchestration         baobab              502e3c07b01f        3 seconds ago       132MB
```

- All Build Options
```shell
build script
Usage:
-------------------------------------------------------------------------------------------------------------------------------------------
  ./build.sh                      : build edge-orchestration by default container
  ./build.sh secure               : build edge-orchestration by default container with secure option
  ./build.sh container            : build Docker container as build system environmet
  ./build.sh container secure     : build Docker container as build system environmet with secure option
  ./build.sh object [Arch]        : build object (c-object, java-object), Arch:{x86, x86_64, arm, arm64} (default:all)
  ./build.sh object secure [Arch] : build object (c-object, java-object) with secure option, Arch:{x86, x86_64, arm, arm64} (default:all)
  ./build.sh clean                : build clean
  ./build.sh test [PKG_NAME]      : run unittests (optional for PKG_NAME)
-------------------------------------------------------------------------------------------------------------------------------------------
```
> If you build the edge-orchestration as c-object, then a more detailed description can be found [x86_64_native.md](x86_64_native.md)
---

## API Document
Edge Orchestration provides REST API for its operation. Description for the APIs are stored in [/doc](../../) folder.
- **[edge_orchestration_api.yaml](../../edge_orchestration_api.yaml)** or 
- **[edge_orchestration_api_secure.yaml](../../edge_orchestration_api_secure.yaml)** for secure mode.

Note that you can visit [Swagger Editor](https://editor.swagger.io/) to graphically investigate the REST API in YAML.

---

## How to work

#### 0. Prerequisites
  - Same network connected among the devices.
  - Same Authentication key in /var/edge-orchestration/user/orchestration_userID.txt
    - Please see the above [4. Add Key file](#4-add-key-file) to know how to add authentication key
  - Edge Orchestration Docker image
    - Please see the above [How to build](#how-to-build) to know how to build Edge Orchestration Docker image


#### 1. Run Edge Orchestration container

```shell
$ docker run -it -d \
                --privileged \
                --network="host" \
                --name edge-orchestration \
                -v /var/edge-orchestration/:/var/edge-orchestration/:rw \
                -v /var/run/docker.sock:/var/run/docker.sock:rw \
                -v /proc/:/process/:ro \
                edge-orchestration:baobab
```
- Result

```shell
$ docker logs -f edge-orchestration

2019/10/16 07:35:45 main_secured.go:89: [interface] OrchestrationInit
2019/10/16 07:35:45 main_secured.go:90: >>> commitID  :  c3041ae
2019/10/16 07:35:45 main_secured.go:91: >>> version   :  
2019/10/16 07:35:45 main_secured.go:92: >>> buildTime :  {build time}
2019/10/16 07:35:45 main_secured.go:93: >>> buildTags :  secure
2019/10/16 07:35:45 discovery.go:256: [discoverymgr] UUID :  {$UUID}
2019/10/16 07:35:45 discovery.go:338: [discoverymgr] [{$discovery_ip_list}]
2019/10/16 07:35:45 discovery.go:369: [deviceDetectionRoutine] edge-orchestration-{$UUID}
2019/10/16 07:35:45 discovery.go:370: [deviceDetectionRoutine] confInfo    : ExecType(container), Platform(docker)
2019/10/16 07:35:45 discovery.go:371: [deviceDetectionRoutine] netInfo     : IPv4({$discovery_ip_list})
2019/10/16 07:35:45 discovery.go:372: [deviceDetectionRoutine] serviceInfo : Services([])
2019/10/16 07:35:45 discovery.go:373: 
2019/10/16 07:35:45 tls.go:40: SetCertFilePath:  /var/edge-orchestration/data/cert
2019/10/16 07:35:45 tls.go:40: SetCertFilePath:  /var/edge-orchestration/data/cert
2019/10/16 07:35:45 route.go:76: {APIV1Ping GET /api/v1/ping 0x8090f0}
2019/10/16 07:35:45 route.go:76: {APIV1ServicemgrServicesPost POST /api/v1/servicemgr/services 0x809160}
2019/10/16 07:35:45 route.go:76: {APIV1ServicemgrServicesNotificationServiceIDPost POST /api/v1/servicemgr/services/notification/{serviceid} 0x8091c0}
2019/10/16 07:35:45 route.go:76: {APIV1ScoringmgrScoreLibnameGet GET /api/v1/scoringmgr/score 0x809220}
2019/10/16 07:35:45 route.go:76: {APIV1RequestServicePost POST /api/v1/orchestration/services 0x806cb0}
2019/10/16 07:35:45 route.go:104: ListenAndServeTLS_For_Inter
2019/10/16 07:35:45 route.go:111: ListenAndServe
2019/10/16 07:35:45 main_secured.go:141: interface orchestration init done
```

#### 2. Request to execute a service

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
  $ curl -X POST "IP:56001/api/v1/orchestration/services" -H "accept: application/json" -H "Content-Type: application/json" -d "{ \"ServiceName\": \"hello-world\", \"ServiceInfo\": [{ \"ExecutionType\": \"container\", \"ExecCmd\": [ \"docker\", \"run\", \"-v\", \"/var/run:/var/run:rw\", \"hello-world\"]}], \"StatusCallbackURI\": \"http://localhost:8888/api/v1/services/notification\"}"
    ```
  ---
   If the `edge-orchestration` was assembled with `secure` option.
   You need to add a JSON Web Token into request header `Authorization: {token}` and a image digest (sha256) to the last parameter. `"hello-world@sha256:fc6a51919cfeb2e6763f62b6d9e8815acbf7cd2e476ea353743570610737b752"`. More information about it you can find [here](../../secure_manager.md).
   ```
   $ curl -X POST "IP:56001/api/v1/orchestration/services" -H "accept: application/json" -H "Content-Type: application/json" -H "Authorization: $EDGE_ORCHESTRATION_TOKEN" -d "{ \"ServiceName\": \"hello-world\", \"ServiceInfo\": [{ \"ExecutionType\": \"container\", \"ExecCmd\": [ \"docker\", \"run\", \"-v\", \"/var/run:/var/run:rw\", \"hello-world@sha256:fc6a51919cfeb2e6763f62b6d9e8815acbf7cd2e476ea353743570610737b752\"]}], \"StatusCallbackURI\": \"http://localhost:8888/api/v1/services/notification\"}"
   ```  
   To add the `EDGE_ORCHESTRATION_TOKEN` variable to the environment execute the next command:
   ```
   $ . tools/jwt_gen.sh HS256
   ```
   To add your container hash to the container white list `/var/edge-erchestration/data/cwl/containerwhitelist.txt`, you need to add a hash line to the end file.  
   ```
   # echo "fc6a51919cfeb2e6763f62b6d9e8815acbf7cd2e476ea353743570610737b752" >> /var/edge-erchestration/data/cwl/containerwhitelist.txt
   ```
  ---

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

