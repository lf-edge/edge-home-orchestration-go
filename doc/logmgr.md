# Logmgr
## Contents
1. [Introduction](#1-introduction)
2. [How to Use](#2-how-to-use)

## 1. Introduction
Logmgr handles the logging of edge-home-orchestration-go project.

## 2. How To Use
To use the logmgr, you should import the logmgr library and allocate the instance into your code like below.
```
import "common/logmgr"
log := logmgr.GetInstance()
```

You can log messages as Info, Error, etc.
```
log.Info("Hello, edge-home-orchestration-go")
```

See [Logrus](https://github.com/sirupsen/logrus) library for more usage.
