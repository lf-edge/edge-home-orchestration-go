# Testing policy
## Contents
1. [Introduction](#1-introduction)
2. [How to start Test Suite (Local)](#2-how-to-start-test-suite-local)  
    2.1 [Using the system build script](#21-using-the-system-build-script)  
    2.2 [Using standard Go language facilities](#22-using-standard-go-language-facilities)  
3. [Automated Run Test Suite (Remote)](#3-automated-run-test-suite-remote)  

---

## 1. Introduction

Testing is a very important part of Edge Orchestration project and our team values it highly.

**When adding or changing functionality, the main requirement is to include new tests as part of your contribution.**
> If your contribution does not have the required tests, please mark this in the PR and our team will support to develop them.

The Edge Orchestration team strives to maintain test coverage of **at least 70%**. We ask you to help us keep this minimum. Additional tests are very welcome.

The Edge Orchestration team strongly recommends adhering to the [Test-driven development (TDD)](https://en.wikipedia.org/wiki/Test-driven_development) as a software development process.

---

## 2. How to start Test Suite (Local)
There are two ways to test:

### 2.1 Using the makefile
To start testing all packages:
```
$ make test
```
To start testing a specific package:
```
$ make test [PKG_NAME]
```

### 2.2 Using standard Go language facilities
To start testing all packages:
```
$ gocov test $(go list ./internal/... | grep -v mock) -coverprofile=/dev/null
```
To start testing a specific package:
```
$ go test -v [PKG_NAME]
```
    
---

## 3. Automated Run Test Suite (Remote)

Code testing occurs remotely using a [github->Actions->workflow (Build)](https://github.com/lf-edge/edge-home-orchestration-go/actions) during each `push` or `pull_request`.

> [More information on github->actions](https://docs.github.com/en/actions) 
---
 
