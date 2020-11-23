# Edge Orchestration on Raspberry Pi 3

[![Raspberry Pi 3](raspberry_pi3.jpg)](https://www.raspberrypi.org/products/raspberry-pi-3-model-b-plus/)

## Preparing Raspberry Pi 3 board

#### 1. Creating an image and writing Raspbian on a SD card
To install the Raspbian operating system follow the [instructions](https://www.raspberrypi.org/documentation/installation/installing-images/README.md).
We recommend using a [balenaEtcher](https://www.balena.io/etcher/) to writing an image of a the Raspbian on the SD card

> SD card must be at least 16 Gb

#### 2. Start Raspberry Pi 3

Insert the SD card into the Raspberry pi 3 and turn on the power. Make configuration settings for the Raspbian the first time you turn it on.

---

## Quick start
This section provides how to download and run pre-built Docker image without building the project.

> TBD

---

## How to build
There are two options for building a edge-orchestration container:
1. On your PC and downloading the edge-orchestration container image from the `edge-orchestration.tar` archive (recommended).
2. Build directly on the Raspberry Pi 3 board.
### 1. Using your PC

Prerequisites: install the qemu packages
```shell
$ sudo apt-get install qemu binfmt-support qemu-user-static
```

Run the `./build.sh` script and specify the build parameters - `container`,  architecture - `arm`  (in the case of building in protected mode, add `secure`), see examples below:
```shell
$ ./build.sh container arm
```
or for protected mode:
```shell
$ ./build.sh container secure arm
```
the build result will be `edge-orchestration.tar` archive that can be found `GoMain/bin/edge-orchestration.tar`

Next, need to copy `edge-orchestration.tar` archive to the Paspberry Pi 3 board, install the docker container (see [here](../x86_64_linux/x86_64_linux.md#Build-Prerequisites) only docker part) and load the image using the command:
```shell
$ docker load -i edge-orchestration.tar
```
The build is finished, how to run see [here](../x86_64_linux/x86_64_linux.md#how-to-work).

### 2. Build directly on the Raspberry Pi 3 board.
#### Build Prerequisites
- docker

```sh
$ curl -sSL https://get.docker.com | sh
$ sudo usermod -aG docker $USER
$ newgrp docker
```

> For [execution of docker commands with non-root privileges](https://docs.docker.com/install/linux/linux-postinstall/#manage-docker-as-a-non-root-user) you need to add `$USER` to docker group.  
`$ sudo usermod -aG docker $USER`

- go compiler (install a version not lower than 1.12.5, recommended v1.14.x)

```sh 
$ wget https://dl.google.com/go/go1.14.5.linux-armv6l.tar.gz
$ sudo tar -C /usr/local -xvf go1.14.5.linux-armv6l.tar.gz
$ export GOPATH=$HOME/go
$ export PATH=$PATH:$GOPATH/bin:/usr/local/go/bin
```

- glide ([package manager for Go](https://glide.readthedocs.io/en/latest/))  

```sh
$ sudo apt install golang-glide
```
> To add 3rd-party packages need to add them to the `glide.yaml` (format is described [here](https://glide.readthedocs.io/en/latest/glide.yaml/))

- edge-orchestration source code

```sh
$ git clone https://github.com/lf-edge/edge-home-orchestration-go.git

```
The build is described [here](../x86_64_linux/x86_64_linux.md#how-to-build).

The build is finished, how to run see [here](../x86_64_linux/x86_64_linux.md#how-to-work).
