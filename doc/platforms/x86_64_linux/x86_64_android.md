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
To build an java-object (`liborchestration.aar/liborchestration-sources.jar`), you must run one of the commands depending on normal/secure mode.
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

[INFO]	Replacing existing vendor dependencies

-----------------------------------
 Build clean
-----------------------------------
go clean
rm -rf /home/virtual-pc/projects/edge-home-orchestration-go/src/interfaces/capi/output
rm -rf /home/virtual-pc/projects/edge-home-orchestration-go/src/interfaces/javaapi/output


**********************************
 Target Binary arch is amd64 
**********************************

----------------------------------------
 Create Static object of Orchestration
----------------------------------------
mkdir -p /home/virtual-pc/projects/edge-home-orchestration-go/src/interfaces/capi/output/inc/linux_x86-64 /home/virtual-pc/projects/edge-home-orchestration-go/src/interfaces/capi/output/lib/linux_x86-64
CGO_ENABLED=1 go build -ldflags '-extldflags "-static" -X main.version= -X main.commitID=70d67d1 -X main.buildTime=20200722.0008 -X main.buildTags=' -o /home/virtual-pc/projects/edge-home-orchestration-go/src/interfaces/capi/output/lib/linux_x86-64/liborchestration.a -buildmode=c-archive interfaces/capi || exit 1
mv /home/virtual-pc/projects/edge-home-orchestration-go/src/interfaces/capi/output/lib/linux_x86-64/liborchestration.h /home/virtual-pc/projects/edge-home-orchestration-go/src/interfaces/capi/output/inc/linux_x86-64/orchestration.h
ls -al /home/virtual-pc/projects/edge-home-orchestration-go/src/interfaces/capi/output/lib/linux_x86-64
total 26764
drwxrwxr-x 2 virtual-pc virtual-pc     4096 лип 22 00:09 .
drwxrwxr-x 3 virtual-pc virtual-pc     4096 лип 22 00:08 ..
-rw-rw-r-- 1 virtual-pc virtual-pc 27397614 лип 22 00:09 liborchestration.a

**********************************
 Target Binary is for Android 
**********************************

-------------------------------------------
 Create Android archive from Java interface
-------------------------------------------
mkdir -p /home/virtual-pc/projects/edge-home-orchestration-go/src/interfaces/javaapi/output
gomobile init
gomobile bind -ldflags '-X main.version= -X main.commitID=70d67d1 -X main.buildTime=20200722.0008 -X main.buildTags=' -o /home/virtual-pc/projects/edge-home-orchestration-go/src/interfaces/javaapi/output/liborchestration.aar -target=android/amd64 -androidapi=23 interfaces/javaapi || exit 1
ls -al /home/virtual-pc/projects/edge-home-orchestration-go/src/interfaces/javaapi/output
total 6316
drwxrwxr-x 2 virtual-pc virtual-pc    4096 лип 22 00:09 .
drwxrwxr-x 3 virtual-pc virtual-pc    4096 лип 22 00:09 ..
-rw-rw-r-- 1 virtual-pc virtual-pc 6444465 лип 22 00:09 liborchestration.aar
-rw-rw-r-- 1 virtual-pc virtual-pc    9996 лип 22 00:09 liborchestration-sources.jar


**********************************
 Edge-orchestration Archive 
**********************************
tree /home/virtual-pc/projects/edge-home-orchestration-go/src/interfaces/capi/output
/home/virtual-pc/projects/edge-home-orchestration-go/src/interfaces/capi/output
├── inc
│   └── linux_x86-64
│       └── orchestration.h
└── lib
    └── linux_x86-64
        └── liborchestration.a

4 directories, 2 files
tree /home/virtual-pc/projects/edge-home-orchestration-go/src/interfaces/javaapi/output
/home/virtual-pc/projects/edge-home-orchestration-go/src/interfaces/javaapi/output
├── liborchestration.aar
└── liborchestration-sources.jar

0 directories, 2 files

-------------------------------------
 Clean up 3rdParty directory
-------------------------------------
rm -rf /home/virtual-pc/projects/edge-home-orchestration-go/vendor
virtual-pc@virtualpc-VirtualBox:~/projects/edge-home-orchestration-go$ ls /home/virtual-pc/projects/edge-home-orchestration-go/src/interfaces/javaapi/output
liborchestration.aar  liborchestration-sources.jar
```

## Example of using java-object (liborchestration.aar/liborchestration-sources.jar)

TBD