# Testing policy
## Contents
1. [Introduction](#1-introduction)
2. [How to start Test Suite (Local)](#2-how-to-start-test-suite-local)  
    2.1 [Using the makefile](#21-using-the-makefile)  
    2.2 [Using standard Go language facilities](#22-using-standard-go-language-facilities)  
3. [Automated Run Test Suite (Remote)](#3-automated-run-test-suite-remote)  
4. [Test file pattern](#4-test-file-pattern)  

---

## 1. Introduction

Testing is a very important part of Edge Orchestration project and our team values it highly.

**When adding or changing functionality, the main requirement is to include new tests as part of your contribution.**
> If your contribution does not have the required tests, please mark this in the PR and our team will support to develop them.

The Edge Orchestration team strives to maintain test coverage of **at least 70%**. We ask you to help us keep this minimum. Additional tests are very welcome.

The Edge Orchestration team strongly recommends adhering to the [Test-driven development (TDD)](https://en.wikipedia.org/wiki/Test-driven_development) as a software development process.

---

## 2. How to start Test Suite (Local)

> Make sure the `gocov` and `gocov-html` packages are installed [(How to install)](https://github.com/matm/gocov-html#installation). Recommendation: Do not install these utilities from the `edge-home-orchestration-go` folder.

There are two ways to test:
### 2.1 Using the makefile
To start testing all packages:
```
make test
```
To start testing a specific package:
```
make test [PKG_NAME]
```

### 2.2 Using standard Go language facilities
To start testing all packages:
```
gocov test $(go list ./internal/... | grep -v mock) -coverprofile=/dev/null
```
To start testing a specific package:
```
go test -v [PKG_NAME]
```

---

## 3. Automated Run Test Suite (Remote)

Code testing occurs remotely using a [github->Actions->workflow (Build)](https://github.com/lf-edge/edge-home-orchestration-go/actions) during each `push` or `pull_request`.

> [More information on github->actions](https://docs.github.com/en/actions) 

---

## 4. Test file pattern

Testing should include both positive (success) and negative (fail) tests whenever possible. Tests can be nested for greater coverage and better understanding. Example:

```
package <your-package-name>

import (
	"testing"
)

const (
	unexpectedSuccess = "unexpected success"
	unexpectedFail    = "unexpected fail"
)

func TestFuncName(t *testing.T) {
	// Positive test
	t.Run("Success", func(t *testing.T) {
		if err := FuncName(param); err != nil {
			t.Error(unexpectedFail)
		}
	})
	// Negative tests
	t.Run("Fail", func(t *testing.T) {
		// Nested test
		t.Run("FailCondition #1", func(t *testing.T) {
			if err := FuncName(param); err == nil {
				t.Error(unexpectedSuccess)
			}
		})
		// Nested test with panic result
		t.Run("FailCondition #2", func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil {
					t.Error(r)
				}
			}()
			if err := FuncName(param); err == nil {
				t.Error(unexpectedSuccess)
			}
		})
	})
}
```
