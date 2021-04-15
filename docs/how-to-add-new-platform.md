# How to add a new platform

# Contents
1. [Introduction](#1-introduction)
2. [Add a new platform](#2-add-a-new-platform)  
    2.1 [Additional code structure](#21-additional-code-structure)  
    2.2 [Update README.md](#22-update-readmemd)  
3. [Code inclusion to Edge-Orchestration upstream](#3-code-inclusion-to-edge-orchestration-upstream)

## 1. Introduction
This document provides a description of how to add a new platform to the Edge-Orchestration project.

## 2. Add a new platform
First of all, you can familiarize yourself with the [Raspberry Pi 3 example](./platforms/raspberry_pi3/raspberry_pi3.md) and do the same.
Porting a new platform is a simple process with three main parts.

### 2.1 Additional code structure

You must first create a directory with the name of your platform `<new platform name folder>`, a file with description `<name_platform>.md` and a picture of the platform `<name_platform>.png`. Example see below:
```
...
├── docs
│   ├── platforms
│   │   ├── <new platform name folder>
│   │   │   ├── <name_platform>.md
│   │   │   └── <name_platform>.png
│   │   ├── raspberry_pi3
│   │   │   ├── raspberry_pi3.jpg
│   │   │   └── raspberry_pi3.md
...

```
#### 2.1.1 <name_platform>.md
The following aspects need to be described in this file:
 * Describe how to create or where to download a Linux image.
 * How to start and configure Linux (network configuration if required).
 * How to download and run pre-built Docker image (`edge-orchestration.tar`) without building and/or describe the procedure for how to build `edge-orchestration.tar` directly on the board.
 * Describe the procedure for launching a docker on the board without root permission.

 >  The above points are required. If you have additional information, you can provide it without hesitation.

#### 2.1.2 <name_platform>.png
The requirement for this file is as follows:
 * Display the board at an angle to make it easy to recognize.
 * Size no more than 150 pixels in width or height.
 * Picture format `png` or `jpg`.
 * Preferably white background.

### 2.2 Update README.md
There is a section [3. Platforms Supported](https://github.com/lf-edge/edge-home-orchestration-go#platforms-supported) which lists all devices that are supported in the Edge-Orchestration, you must add your platform by specifying a link to the description and picture of the platform.
You also need to add a platform description into [Quick start guides for supported platforms](https://github.com/lf-edge/edge-home-orchestration-go#quick-start-guides-for-supported-platforms)
where:
 * **Platform** - platform name and link to quick start guide
 * **Maintained** - version or repository tag the Edge-Orchestration was tested latest
 * **Maintainer** - contact person who can help to run Edge-Orchestration on the this platform
 * **Remarks** - additional important information

## 3. Code inclusion to Edge-Orchestration upstream
We do encourage everyone to submit their board support to the edge-orchestration project itself, so it becomes part of the official releases and will be maintained by the edge-orchestration community itself. If you intend to do so, then there are a few more things that you are supposed to do.

If you are submitting the board support upstream and cannot give Edge-Orchestration maintainers your platform, then we are going to ask you to become the maintainer of the platform you have added. By being a maintainer for the platform you are responsible to keep it up to date and we will be ask to test every the Edge-Orchestration release on your platform.

[README.md]: ../README.md
