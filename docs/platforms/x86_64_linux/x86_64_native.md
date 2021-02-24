# Edge Orchestration on x86_64 (TypeExec: native)
The edge-orchestration can be launched as a native Linux application and run services as native Linux processes (outside docker).

## How to build
The general preparation steps are described [here](x86_64_linux.md#How-to-build).
To build an с-object (liborchestration.a), you must run one of the commands depending on normal/secure mode.
```
...
./build.sh object [Arch]        : build object (c-object, java-object), Arch:{x86, x86_64, arm, arm64} (default:all)
./build.sh object secure [Arch] : build object (c-object, java-object) with secure option, Arch:{x86, x86_64, arm, arm64} (default:all)
...
```
for example:
```
$ ./build.sh object x86_64
...
**********************************
 Target Binary arch is amd64 
**********************************

----------------------------------------
 Create Static object of Orchestration
----------------------------------------
mkdir -p /home/virtual-pc/projects/edge-home-orchestration-go/bin/capi/output/inc/linux_x86-64 /home/virtual-pc/projects/edge-home-orchestration-go/bin/capi/output/lib/linux_x86-64
CGO_ENABLED=1 GO111MODULE=on go build -ldflags '-extldflags "-static" -X main.version= -X main.commitID=687e09c -X main.buildTime=20210213.0901 -X main.buildTags=' -o /home/virtual-pc/projects/edge-home-orchestration-go/bin/capi/output/lib/linux_x86-64/liborchestration.a -buildmode=c-archive /home/virtual-pc/projects/edge-home-orchestration-go/cmd/edge-orchestration/capi || exit 1
mv /home/virtual-pc/projects/edge-home-orchestration-go/bin/capi/output/lib/linux_x86-64/liborchestration.h /home/virtual-pc/projects/edge-home-orchestration-go/bin/capi/output/inc/linux_x86-64/orchestration.h
ls -al /home/virtual-pc/projects/edge-home-orchestration-go/bin/capi/output/lib/linux_x86-64
total 37100
drwxrwxr-x 2 virtual-pc virtual-pc     4096 Feb 13 09:01 .
drwxrwxr-x 3 virtual-pc virtual-pc     4096 Feb 13 09:01 ..
-rw-rw-r-- 1 virtual-pc virtual-pc 37980926 Feb 13 09:01 liborchestration.a
```
## Example of using c-object (liborchestration.c)
The example uses the `ls` command instead of a service.
> It should be noted that you must ensure the visibility of your service from any point (for example, by copying it to the `/bin` folder or add to `PATH`)
Source code structure below:
```
examples
└── native
    ├── main.c
    └── Makefile
test
└── native
    ├── copy_srvs.sh
    └── ls_srv
        └── ls_srv.conf
tools
└── create_fs.sh 

```

1. To start for the first time, need to create a file system (only once need to execute the command)
```
$ cd tools
$ sudo ./create_fs.sh
```

2. Copy the folder with the service configuration file from `services` to `/var/edge-orchestration/apps` or use `copy_srvs.sh` to copy all service folders to `/var/edge-orchestration/apps`.
 ```
 $ cd test/native
 $ sudo ./copy_srvs.sh
```
> The structure of the [configuration file](../../../internal/controller/configuremgr/native/description/doc.go) and example can be found [ls_srv.conf](../../../test/native/ls_srv/ls_srv.conf).

3. To build the native edge-orchestration, run the following commands:
```
$ cd examples/native
$ make
CC: main.c
gcc -c -I../../bin/capi/output/inc/linux_x86-64 main.c -o main.o
gcc   main.o -L../../bin/capi/output/lib/linux_x86-64 -pthread -lorchestration -o edge-orchestration

```
4. Run native edge-orchestration
```
$ sudo ./edge-orchestration 
2020/07/20 09:24:10 main.go:158: [interface] OrchestrationInit
2020/07/20 09:24:10 main.go:159: >>> commitID  :  094ca91
2020/07/20 09:24:10 main.go:160: >>> version   :  
2020/07/20 09:24:10 main.go:161: >>> buildTime :  20200720.0832
2020/07/20 09:24:10 main.go:162: >>> buildTags :  
2020/07/20 09:24:10 discovery.go:257: [discoverymgr] UUID :  1da15e3d-09d4-4f80-ad72-6ca943dd5bcf
2020/07/20 09:24:11 helper.go:99: [http://10.0.2.15:56002/api/v1/ping] reqeust get failed !!, err = Get "http://10.0.2.15:56002/api/v1/ping": dial tcp 10.0.2.15:56002: connect: connection refused
Get "http://10.0.2.15:56002/api/v1/ping": dial tcp 10.0.2.15:56002: connect: connection refused
2020/07/20 09:24:11 discovery.go:339: [discoverymgr] [10.0.2.15]
2020/07/20 09:24:11 discovery.go:370: [deviceDetectionRoutine] edge-orchestration-1da15e3d-09d4-4f80-ad72-6ca943dd5bcf
2020/07/20 09:24:11 discovery.go:371: [deviceDetectionRoutine] confInfo    : ExecType(native), Platform(linux)
2020/07/20 09:24:11 discovery.go:372: [deviceDetectionRoutine] netInfo     : IPv4([10.0.2.15]), RTT(0)
2020/07/20 09:24:11 discovery.go:373: [deviceDetectionRoutine] serviceInfo : Services([])
2020/07/20 09:24:11 discovery.go:374: 
2020/07/20 09:24:11 native.go:176: [configuremgr] confPath : /var/edge-orchestration/apps/ls_srv/ls_srv.conf
2020/07/20 09:24:11 native.go:144: [configuremgr] ServiceName: ls
2020/07/20 09:24:11 native.go:145: [configuremgr] ExecutableFileName: ls
2020/07/20 09:24:11 native.go:146: [configuremgr] AllowedRequester: [bash]
2020/07/20 09:24:11 discovery.go:435: [discoverymgr] [mdnsTXTSizeChecker] size ::  13  Bytes
2020/07/20 09:24:11 native.go:125: start watching for /var/edge-orchestration/apps
2020/07/20 09:24:11 native.go:126: configuremgr watcher register end
2020/07/20 09:24:11 route.go:78: {APIV1Ping GET /api/v1/ping 0x7a8280}
2020/07/20 09:24:11 route.go:78: {APIV1ServicemgrServicesPost POST /api/v1/servicemgr/services 0x7a82f0}
2020/07/20 09:24:11 route.go:78: {APIV1ServicemgrServicesNotificationServiceIDPost POST /api/v1/servicemgr/services/notification/{serviceid} 0x7a8350}
2020/07/20 09:24:11 route.go:78: {APIV1ScoringmgrScoreLibnameGet GET /api/v1/scoringmgr/score 0x7a83b0}
2020/07/20 09:24:11 route.go:109: ListenAndServe_For_Inter
2020/07/20 09:24:11 route.go:113: ListenAndServe
2020/07/20 09:24:11 main.go:226: interface orchestration init done
2020/07/20 09:24:11 main.go:233: [interface] OrchestrationRequestService
2020/07/20 09:24:11 main.go:262: [OrchestrationRequestService] appName:ls
2020/07/20 09:24:11 main.go:263: [OrchestrationRequestService] selfSel:true
2020/07/20 09:24:11 main.go:264: [OrchestrationRequestService] requester:bash
2020/07/20 09:24:11 main.go:265: [OrchestrationRequestService] infos:[{native [ls]}]
2020/07/20 09:24:11 orchestration_api.go:122: [RequestService] ls: [{native [ls]}]
2020/07/20 09:24:11 orchestration_api.go:146: [RequestService] getCandidate
2020/07/20 09:24:11 orchestration_api.go:148: [0] Id       : edge-orchestration-1da15e3d-09d4-4f80-ad72-6ca943dd5bcf
2020/07/20 09:24:11 orchestration_api.go:149: [0] ExecType : native
2020/07/20 09:24:11 orchestration_api.go:150: [0] Endpoint : [10.0.2.15]
2020/07/20 09:24:11 orchestration_api.go:151: 
2020/07/20 09:24:11 orchestration_api.go:316: [orchestrationapi] deviceScore
2020/07/20 09:24:11 orchestration_api.go:317: candidate Id       : edge-orchestration-1da15e3d-09d4-4f80-ad72-6ca943dd5bcf
2020/07/20 09:24:11 orchestration_api.go:318: candidate ExecType : native
2020/07/20 09:24:11 orchestration_api.go:319: candidate Endpoint : 10.0.2.15
2020/07/20 09:24:11 orchestration_api.go:320: candidate score    : 5.238915966812347
2020/07/20 09:24:11 orchestration_api.go:222: [orchestrationapi]  [{edge-orchestration-1da15e3d-09d4-4f80-ad72-6ca943dd5bcf 10.0.2.15 5.238915966812347 native}]
2020/07/20 09:24:11 main.go:273: requestService handle :  {ERROR_NONE ls {native 10.0.2.15}}
2020/07/20 09:24:11 nativeexecutor.go:57: [nativeexecutor] ls [ls]
2020/07/20 09:24:11 nativeexecutor.go:58: [nativeexecutor] parameter length : 1
2020/07/20 09:24:11 main.go:320: Message: ERROR_NONE
2020/07/20 09:24:11 main.go:320: ServiceName: ls
2020/07/20 09:24:11 main.go:320: ExecutionType: native
2020/07/20 09:24:11 main.go:320: Target: 10.0.2.15
2020/07/20 09:24:11 nativeexecutor.go:120: edge-orchestration
2020/07/20 09:24:11 nativeexecutor.go:120: main.c
2020/07/20 09:24:11 nativeexecutor.go:120: main.o
2020/07/20 09:24:11 nativeexecutor.go:120: Makefile
2020/07/20 09:24:11 nativeexecutor.go:65: [nativeexecutor] Just ran subprocess  7216
2020/07/20 09:24:11 nativeexecutor.go:141: [nativeexecutor] ls is exited with no error
2020/07/20 09:24:11 orchestration_api.go:342: [orchestrationapi] service status changed [appNames:ls][status:Finished]
2020/07/20 09:24:11 discovery.go:370: [deviceDetectionRoutine] edge-orchestration-1da15e3d-09d4-4f80-ad72-6ca943dd5bcf
2020/07/20 09:24:11 discovery.go:371: [deviceDetectionRoutine] confInfo    : ExecType(native), Platform(linux)
2020/07/20 09:24:11 discovery.go:372: [deviceDetectionRoutine] netInfo     : IPv4([]), RTT(0)
2020/07/20 09:24:11 discovery.go:373: [deviceDetectionRoutine] serviceInfo : Services([])
2020/07/20 09:24:11 discovery.go:374: 
```

## REST API
Edge Orchestration provides external REST API for execute service.

#### Request to execute a service

REST API
- POST  
- **IP:56001/api/v1/orchestration/services** 
- BODY : 
    ```json
    {
        "ServiceRequester": "curl",
        "ServiceName": "ls",
        "ServiceInfo": [
        {
            "ExecutionType": "native",
            "ExecCmd": [
                "ls"
            ]
        }],
        "StatusCallbackURI": "http://localhost:8888/api/v1/services/notification"
    }
    ```
- Curl Example:
```json
$ curl -X POST "IP:56001/api/v1/orchestration/services" -H "accept: application/json" -H "Content-Type: application/json" -d "{ \"ServiceRequester\": \"curl\", \"ServiceName\": \"ls\", \"ServiceInfo\": [{ \"ExecutionType\": \"native\", \"ExecCmd\": [ \"ls\"]}], \"StatusCallbackURI\": \"http://localhost:8888/api/v1/services/notification\"}"
```
Response:
```
{"Message":"ERROR_NONE","RemoteTargetInfo":{"ExecutionType":"native","Target":"10.0.2.15"},"ServiceName":"ls"}
```
log:
```
2020/07/28 17:07:59 externalhandler.go:89: [RestExternalInterface] APIV1RequestServicePost
2020/07/28 17:07:59 externalhandler.go:153: port:  50494
2020/07/28 17:07:59 externalhandler.go:158: requester:  
2020/07/28 17:07:59 orchestration_api.go:122: [RequestService] ls: [{native [ls]}]
2020/07/28 17:07:59 orchestration_api.go:146: [RequestService] getCandidate
2020/07/28 17:07:59 orchestration_api.go:148: [0] Id       : edge-orchestration-1da15e3d-09d4-4f80-ad72-6ca943dd5bcf
2020/07/28 17:07:59 orchestration_api.go:149: [0] ExecType : native
2020/07/28 17:07:59 orchestration_api.go:150: [0] Endpoint : [10.0.2.15]
2020/07/28 17:07:59 orchestration_api.go:151: 
2020/07/28 17:07:59 orchestration_api.go:316: [orchestrationapi] deviceScore
2020/07/28 17:07:59 orchestration_api.go:317: candidate Id       : edge-orchestration-1da15e3d-09d4-4f80-ad72-6ca943dd5bcf
2020/07/28 17:07:59 orchestration_api.go:318: candidate ExecType : native
2020/07/28 17:07:59 orchestration_api.go:319: candidate Endpoint : 10.0.2.15
2020/07/28 17:07:59 orchestration_api.go:320: candidate score    : 20.721974414135058
2020/07/28 17:07:59 orchestration_api.go:222: [orchestrationapi]  [{edge-orchestration-1da15e3d-09d4-4f80-ad72-6ca943dd5bcf 10.0.2.15 20.721974414135058 native}]
2020/07/28 17:07:59 route.go:125: From [127.0.0.1:50494] POST /api/v1/orchestration/services APIV1RequestServicePost 3.645139ms
2020/07/28 17:07:59 nativeexecutor.go:57: [nativeexecutor] ls [ls]
2020/07/28 17:07:59 nativeexecutor.go:58: [nativeexecutor] parameter length : 1
2020/07/28 17:07:59 nativeexecutor.go:120: edge-orchestration
2020/07/28 17:07:59 nativeexecutor.go:120: main.c
2020/07/28 17:07:59 nativeexecutor.go:120: main.o
2020/07/28 17:07:59 nativeexecutor.go:120: Makefile
2020/07/28 17:07:59 nativeexecutor.go:65: [nativeexecutor] Just ran subprocess  11653
2020/07/28 17:07:59 nativeexecutor.go:141: [nativeexecutor] ls is exited with no error
2020/07/28 17:07:59 orchestration_api.go:342: [orchestrationapi] service status changed [appNames:ls][status:Finished]
```
---
If the `edge-orchestration` was assembled with `secure` option.
You need to add a JSON Web Token into request header `Authorization: {token}`. More information about it you can find [here](../../secure_manager.md).

> To add the `EDGE_ORCHESTRATION_TOKEN` variable to the environment execute the next command:
```
$ . tools/jwt_gen.sh HS256 Admin
```

```
$ curl -X POST "127.0.0.1:56001/api/v1/orchestration/services" -H "accept: application/json" -H "Content-Type: application/json" -H "Authorization: $EDGE_ORCHESTRATION_TOKEN" -d "{ \"ServiceRequester\": \"curl\", \"ServiceName\": \"ls\", \"ServiceInfo\": [{ \"ExecutionType\": \"native\", \"ExecCmd\": [ \"ls\"]}], \"StatusCallbackURI\": \"http://localhost:8888/api/v1/services/notification\"}"
```
Response:
```
{"Message":"ERROR_NONE","RemoteTargetInfo":{"ExecutionType":"native","Target":"10.0.2.15"},"ServiceName":"ls"}
```  
