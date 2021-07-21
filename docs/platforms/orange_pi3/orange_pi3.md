# Edge Orchestration on Orange Pi 3

[![Orange Pi 3](orange_pi3.jpg)](http://www.orangepi.org/Orange%20Pi%203/)

## Preparing Orange Pi 3 board

#### 1. Creating an image and writing Orangepi on a SD card
1. Download Ubuntu image for Orange Pi 3 from (http://www.orangepi.org/downloadresources/).
2. To install the Orangepi operating system follow the [instructions](http://www.orangepi.org/Docs/SDcardinstallation.html).

> SD card must be at least 16 Gb

#### 2. Start Orange Pi 3

Insert the SD card into the Orange Pi 3 and turn on the power. Make configuration settings for the Orange the first time you turn it on.

---

## Quick start
This section provides how to download and run pre-built Docker image without building the project.

> TBD

---

## How to build
There are two options for building a edge-orchestration container:
1. On your PC and downloading the edge-orchestration container image from the `edge-orchestration.tar` archive (recommended).
2. Build directly on the Orange Pi 3 board.
### 1. Using your PC

Prerequisites: install the qemu packages
```shell
$ sudo apt-get install qemu binfmt-support qemu-user-static
```

Run the `make create_context` and specify the configuration file name `arm64c` and `make` (in the case of building in protected mode, use add `arm64cs`), see examples below:
```
$ make distclean
$ make create_context CONFIGFILE=arm64c
$ make
```
or for protected mode:
```shell
$ make distclean
$ make create_context CONFIGFILE=arm64cs
$ make
```

> To change the configuration file, you must execute the command `make distclean`

The build result will be `edge-orchestration.tar` archive that can be found `bin/edge-orchestration.tar`

Next, need to copy `edge-orchestration.tar` archive to the Paspberry Pi 3 board, install the docker container (see [here](../x86_64_linux/x86_64_linux.md#Build-Prerequisites) only docker part) and load the image using the command:
```shell
$ docker load -i edge-orchestration.tar
```
The build is finished, how to run see [here](../x86_64_linux/x86_64_linux.md#how-to-work). If you find the issue with Curl Example then please reboot board and run Edge Orchestration container again.

### 2. Build directly on the Orange Pi 3 board
#### Build Prerequisites
- docker

```sh
$ sudo update-alternatives --set iptables /usr/sbin/iptables-legacy
$ sudo update-alternatives --set ip6tables /usr/sbin/ip6tables-legacy
$ curl -sSL https://get.docker.com | sh
$ sudo usermod -aG docker $USER
$ newgrp docker
```

> For [execution of docker commands with non-root privileges](https://docs.docker.com/install/linux/linux-postinstall/#manage-docker-as-a-non-root-user) you need to add `$USER` to docker group.  
`$ sudo usermod -aG docker $USER`

- go compiler (install a version not lower than 1.12.5)

```sh
$ wget https://dl.google.com/go/go1.16.2.linux-arm64.tar.gz
$ tar -C $HOME -xvf go1.16.2.linux-arm64.tar.gz
$ export GOPATH=$HOME/go
$ export PATH=$PATH:$GOPATH/bin
```

- edge-orchestration source code

```sh
$ git clone https://github.com/lf-edge/edge-home-orchestration-go.git

```
The build is described [here](../x86_64_linux/x86_64_linux.md#how-to-build).

The build is finished, how to run see [here](../x86_64_linux/x86_64_linux.md#how-to-work).
