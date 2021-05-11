package libcurl

/*
#cgo CFLAGS: -I./include/
#cgo darwin LDFLAGS: -framework Security -framework Foundation -L${SRCDIR}/darwin -lcurl -lcrypto -lquiche -lssl
#cgo linux  LDFLAGS: -Wl,--no-as-needed -ldl -L${SRCDIR}/linux -lcurl -lcrypto -lquiche -lssl -lz

#include <stdlib.h>
#include <string.h>
#include "./include/curl.h"
*/
import "C"

import (
	"unsafe"
)

//export goCallHeaderFunction
func goCallHeaderFunction(ptr *C.char, size C.size_t, ctx unsafe.Pointer) uintptr {
	curl := context_map.Get(uintptr(ctx))
	buf := C.GoBytes(unsafe.Pointer(ptr), C.int(size))
	if curl != nil && (*curl.headerFunction)(buf, curl.headerData) {
		return uintptr(size)
	}
	return C.CURL_WRITEFUNC_PAUSE
}

//export goCallWriteFunction
func goCallWriteFunction(ptr *C.char, size C.size_t, ctx unsafe.Pointer) uintptr {
	curl := context_map.Get(uintptr(ctx))
	buf := C.GoBytes(unsafe.Pointer(ptr), C.int(size))
	if curl != nil && (*curl.writeFunction)(buf, curl.writeData) {
		return uintptr(size)
	}
	return C.CURL_WRITEFUNC_PAUSE
}

//export goCallProgressFunction
func goCallProgressFunction(dltotal, dlnow, ultotal, ulnow C.double, ctx unsafe.Pointer) int {
	curl := context_map.Get(uintptr(ctx))
	if curl != nil && (*curl.progressFunction)(float64(dltotal), float64(dlnow),
		float64(ultotal), float64(ulnow),
		curl.progressData) {
		return 0
	}
	return 1
}

//export goCallReadFunction
func goCallReadFunction(ptr *C.char, size C.size_t, ctx unsafe.Pointer) uintptr {
	curl := context_map.Get(uintptr(ctx))
	buf := C.GoBytes(unsafe.Pointer(ptr), C.int(size))
	ret := (*curl.readFunction)(buf, curl.readData)
	str := C.CString(string(buf))
	defer C.free(unsafe.Pointer(str))
	if curl != nil && C.memcpy(unsafe.Pointer(ptr), unsafe.Pointer(str), C.size_t(ret)) == nil {
		panic("read_callback memcpy error!")
	}
	return uintptr(ret)
}

//export goCallDebugFunction
func goCallDebugFunction(ptr *C.char, size C.size_t, ctx unsafe.Pointer) int {
	curl := context_map.Get(uintptr(ctx))
	data := C.GoBytes(unsafe.Pointer(ptr), C.int(size))
        if curl != nil && curl.debugFunction != nil {
        	return (*curl.debugFunction)(data, curl.debugData)
        } else {
		return 0
        }
}
