# Edge Orchestration on x86_64 (TypeExec: android)

## Prerequisites:
- Android SDK
1. Download [Android SDK](https://developer.android.com/studio/index.html#command-tools), then extract the downloaded package.
2. Wrapping `tools` directory inside `cmdline-tools`
3. Go to `cmdline-tools/tools/bin` folder and execute following commands: 
```
$ ./sdkmanager "platform-tools" "platforms;android-28"
$ export ANDROID_SDK=<path to Android SDK directory>
$ export ANDROID_HOME=<path to Android SDK directory>
$ export PATH=$PATH:$ANDROID_SDK/cmdline-tools/tools:$ANDROID_SDK/platform-tools
```
- Android NDK
1. Go to `cmdline-tools/tools/bin` folder and execute following commands:
```
$ ./sdkmanager --list
```
Select NDK version from the list (recommended `ndk;21.3.6528147`)
```
$ ./sdkmanager --install "ndk;21.3.6528147"
$ export ANDROID_NDK_HOME=<path to Android NDK directory> 
$ export PATH=$PATH:$ANDROID_NDK_HOME
```
> `ndk;22.x` is temporarily not supported.

## How to build
The general preparation steps are described [here](x86_64_linux.md).
To build an java-object (`liborchestration.aar/liborchestration-sources.jar`), you must run commands depending on configuration file.

Run the `make create_context` and specify the configuration file name `x86_64a` and `make` (in the case of building in protected mode, use add `x86_64as`), see examples below:
```
$ make distclean
$ make create_context CONFIGURATION_FILE_NAME=x86_64a
$ make
```
or for protected mode:
```shell
$ make distclean
$ make create_context CONFIGURATION_FILE_NAME=x86_64as
$ make
```
```
-----------------------------------
 Build clean
-----------------------------------
GO111MODULE=on go clean
rm -rf /home/virtual-pc/projects/edge-home-orchestration-go/vendor
rm -rf /home/virtual-pc/projects/edge-home-orchestration-go/bin/capi/output
rm -rf /home/virtual-pc/projects/edge-home-orchestration-go/bin/javaapi/output

-----------------------------------
 Go Mod Vendor
-----------------------------------
GO111MODULE=on go mod vendor

**********************************
 Target Binary is for Android 
**********************************

-------------------------------------------
 Create Android archive from Java interface
-------------------------------------------
mkdir -p /home/virtual-pc/projects/edge-home-orchestration-go/bin/javaapi/output
rm -rf /home/virtual-pc/projects/edge-home-orchestration-go/vendor
gomobile bind -ldflags '-X main.version= -X main.commitID=687e09c -X main.buildTime=20210213.0915 -X main.buildTags=' -o /home/virtual-pc/projects/edge-home-orchestration-go/bin/javaapi/output/liborchestration.aar -target=android/amd64 -androidapi=23 /home/virtual-pc/projects/edge-home-orchestration-go/cmd/edge-orchestration/javaapi || exit 1
ls -al /home/virtual-pc/projects/edge-home-orchestration-go/bin/javaapi/output
total 8368
drwxrwxr-x 2 virtual-pc virtual-pc    4096 Feb 13 09:16 .
drwxrwxr-x 3 virtual-pc virtual-pc    4096 Feb 13 09:15 ..
-rw-rw-r-- 1 virtual-pc virtual-pc 8544402 Feb 13 09:16 liborchestration.aar
-rw-rw-r-- 1 virtual-pc virtual-pc   10421 Feb 13 09:16 liborchestration-sources.jar


**********************************
 Edge-orchestration Archive 
**********************************
tree /home/virtual-pc/projects/edge-home-orchestration-go/bin/capi/output
/home/virtual-pc/projects/edge-home-orchestration-go/bin/capi/output
├── inc
│   └── linux_x86-64
│       └── orchestration.h
└── lib
    └── linux_x86-64
        └── liborchestration.a

4 directories, 2 files
tree /home/virtual-pc/projects/edge-home-orchestration-go/bin/javaapi/output
/home/virtual-pc/projects/edge-home-orchestration-go/bin/javaapi/output
├── liborchestration.aar
└── liborchestration-sources.jar

0 directories, 2 files
```

## Example of using java-object (liborchestration.aar/liborchestration-sources.jar)

> TBD