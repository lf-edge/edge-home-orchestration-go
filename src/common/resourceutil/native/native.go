/*******************************************************************************
 * Copyright 2019 Samsung Electronics All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 *******************************************************************************/

// Package native implements specific resourceutil functions for LinuxOS
package native

/*
#include <stdlib.h>
#include <dlfcn.h>
#cgo LDFLAGS: -ldl

extern int cgo_getResource(char *name, double* out);

static int getResource(char* name, double* out) {
	return cgo_getResource(name, out);
}

static double wrapper(void* ptr) {
	double (*vptr)(int (*)(char* name, double* out)) = (double (*)(int (*)(char* name, double* out)))ptr;
	return vptr(getResource);
}

*/
import "C"
import (
	"unsafe"

	errs "common/errors"
	"common/resourceutil"
)

// Getter has pointers of dynamic libarary for specific service application
type Getter struct {
	Dl     unsafe.Pointer
	Symbol unsafe.Pointer
}

var resourceIns resourceutil.GetResource

const (
	errNone = 0
	invalidParamError
	systemError
	notSupportedError
)

func init() {
	resourceIns = &resourceutil.ResourceImpl{}
}

// Run gets the infomations of device resource usage
func (t Getter) Run(ID string) float64 {
	resourceIns.SetDeviceID(ID)
	ret := C.wrapper(t.Symbol)
	return float64(ret)
}

// Close stops device resource monitoring
func (t Getter) Close() {
	C.dlclose(t.Dl)
}

//export cgo_getResource
func cgo_getResource(name *C.char, out *C.double) C.int {
	value, err := resourceIns.GetResource(C.GoString(name))
	*out = (C.double)(value)
	return C.int(errorCovertToEnum(err))
}

func errorCovertToEnum(err error) (ret int) {
	switch err.(type) {
	case errs.InvalidParam:
		ret = invalidParamError
	case errs.SystemError:
		ret = systemError
	case errs.NotSupport:
		ret = notSupportedError
	}
	return
}
