# Secure Manager
## Contents
1. [Introduction](#1-introduction)
2. [Verifier](#2-verifier)  
    2.1 [Description](#21-description)      
    2.2 [Workflow](#22-workflow)      
    2.3 [Verifier Management](#23-verifier-management)     
    2.4 [Usage Edge-Orchestration with Verifier](#24-usage-edge-orchestration-with-verifier)

## 1. Introduction
The **Secure Manager** is designed to control security components. Currently it is in an initial state of development and includes the following components:
  1. Verifier  
  2. TBD  
 
In order for the Secure Manager to be allowed, it is necessary to assemble the **Edge-Orchestration** with the `secure` option.

---

## 2. Verifier
### 2.1 Description
The verifier ensures that only allowed containers (images) are launched; a list of allowed containers (their sha256 hashes) is stored in a `/var/edge-orchestration/data/cwl/containerwhitelist.txt` file.

### 2.2 Workflow
 > TBD
---
 
### 2.3 Verifier Management
There are two ways to change the container white list:
1. Using REST API
2. Editing the `containerwhitelist.txt` file (root rights required)

#### 2.3.1 REST API
 REST API supports the next functions for contaner white list management:
 * _**addHashCWL**_ - adds container image hash to the container white list
 * _**delHashCWL**_ - deletes container image hash from the container white list
 * _**delAllHashCWL**_ - deletes all container image hashes from container white list
 * _**printAllHashCWL**_ - displays all hashes (from the container white list) in the log.

Examples of using these commands are given below:
 - POST  
 - **IP:56001/api/v1/orchestration/securemgr** 
 - BODY (depends on the command): 
 
_**addHashCWL**_  
JSON:
```json
    {
        "SecCompName": "Verifier",
        "TypeCmd": "addHashCWL",
        "Desc": [
        {
            "ContainerHash": "fc6a51919cfeb2e6763f62b6d9e8815acbf7cd2e476ea353743570610737b752"
        },
        {
            "ContainerHash": "fc6a51919cfeb2e6763f62b6d9e8815acbf7cd2e476ea353743570610737b751"
        }],
        "StatusCallbackURI": "http://localhost:8888/api/v1/services/notification"
    }
```
Curl:
```shell
curl -X POST "127.0.0.1:56001/api/v1/orchestration/securemgr" -H "accept: application/json" -H "Content-Type: application/json" -d "{ \"SecureMgr\": \"Verifier\", \"CmdType\": \"addHashCWL\", \"Desc\": [{ \"ContainerHash\": \"fc6a51919cfeb2e6763f62b6d9e8815acbf7cd2e476ea353743570610737b752\"}, { \"ContainerHash\": \"fc6a51919cfeb2e6763f62b6d9e8815acbf7cd2e476ea353743570610737b751\"}], \"StatusCallbackURI\": \"http://localhost:8888/api/v1/services/notification\"}"
```

_**delHashCWL**_  
JSON:
```json
    {
        "SecCompName": "Verifier",
        "TypeCmd": "delHashCWL",
        "Desc": [
        {
            "ContainerHash": "fc6a51919cfeb2e6763f62b6d9e8815acbf7cd2e476ea353743570610737b752"
        }],
        "StatusCallbackURI": "http://localhost:8888/api/v1/services/notification"
    }
```
Curl:
```shell
curl -X POST "127.0.0.1:56001/api/v1/orchestration/securemgr" -H "accept: application/json" -H "Content-Type: application/json" -d "{ \"SecureMgr\": \"Verifier\", \"CmdType\": \"delHashCWL\", \"Desc\": [{ \"ContainerHash\": \"fc6a51919cfeb2e6763f62b6d9e8815acbf7cd2e476ea353743570610737b752\"}], \"StatusCallbackURI\": \"http://localhost:8888/api/v1/services/notification\"}"
```
_**delAllHashCWL**_  
JSON:
```json
    {
        "SecCompName": "Verifier",
        "TypeCmd": "delAllHashCWL",
        "StatusCallbackURI": "http://localhost:8888/api/v1/services/notification"
    }
```
Curl:
```shell
curl -X POST "127.0.0.1:56001/api/v1/orchestration/securemgr" -H "accept: application/json" -H "Content-Type: application/json" -d "{ \"SecureMgr\": \"Verifier\", \"CmdType\": \"delAllHashCWL\", \"StatusCallbackURI\": \"http://localhost:8888/api/v1/services/notification\"}"
```
_**printAllHashCWL**_  
JSON:
```json
    {
        "SecCompName": "Verifier",
        "TypeCmd": "printAllHashCWL",
        "StatusCallbackURI": "http://localhost:8888/api/v1/services/notification"
    }
```
Curl:
```shell
curl -X POST "127.0.0.1:56001/api/v1/orchestration/securemgr" -H "accept: application/json" -H "Content-Type: application/json" -d "{ \"SecureMgr\": \"Verifier\", \"CmdType\": \"printAllHashCWL\", \"StatusCallbackURI\": \"http://localhost:8888/api/v1/services/notification\"}"
```

#### 2.3.2 Editing the container white list by other tools.
  
The `/var/edge-orchestration/data/cwl/containerwhitelist.txt` file consists of records that include: container's image sha256 hash (64-ASCII symbols) + '\n'
Therefore, it can be edited with any editor or from the command line.
Example: _how to add hash by command line_
```shell
# echo "fc6a51919cfeb2e6763f62b6d9e8815acbf7cd2e476ea353743570610737b752" >> /var/edge-erchestration/data/cwl/containerwhitelist.txt
```
---

### 2.4 Usage Edge-Orchestration with Verifier
To run **Edge Orchestration** container you need to add a digest (sha256) to the last parameter. For example:  `"hello-world@sha256:fc6a51919cfeb2e6763f62b6d9e8815acbf7cd2e476ea353743570610737b752"`
```
$ curl -X POST "IP:56001/api/v1/orchestration/services" -H "accept: application/json" -H "Content-Type: application/json" -d "{ \"ServiceName\": \"hello-world\", \"ServiceInfo\": [{ \"ExecutionType\": \"container\", \"ExecCmd\": [ \"docker\", \"run\", \"-v\", \"/var/run:/var/run:rw\", \"hello-world@sha256:fc6a51919cfeb2e6763f62b6d9e8815acbf7cd2e476ea353743570610737b752\"]}], \"StatusCallbackURI\": \"http://localhost:8888/api/v1/services/notification\"}"
```  
If the `"fc6a51919cfeb2e6763f62b6d9e8815acbf7cd2e476ea353743570610737b752"` hash is written to the `/var/edge-erchestration/data/cwl/containerwhitelist.txt` file, the container will be launched successfully.

---