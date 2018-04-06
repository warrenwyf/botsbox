package webkit

/*
#cgo pkg-config: webkit2gtk-4.0

#include "gasyncreadycallback.go.h"
#include <webkit2/webkit2.h>
*/
import "C"

import (
	"errors"
	"unsafe"
)

var asyncReadyCallback = C.GAsyncReadyCallback(C.my_g_async_ready_callback)

//export goGAsyncReadyCallback
func goGAsyncReadyCallback(result unsafe.Pointer, userData C.gpointer) {
	cCallbackWrapper := (*C.callback_wrapper)(unsafe.Pointer(userData))

	defer func() {
		cCallbackWrapper.fn = nil
		C.free_callback_wrapper(cCallbackWrapper) // Free C.make_callback_wrapper()
	}()

	fnPtr := (*func(*C.GAsyncResult))(cCallbackWrapper.fn)
	cResult := (*C.GAsyncResult)(result)

	(*fnPtr)(cResult)
}

func makeCallbackCgo(fn func(*C.GAsyncResult)) (C.gpointer, unsafe.Pointer, error) {
	if fn == nil {
		return nil, nil, errors.New("Callback can not be nil")
	}

	cCallbackWrapper := C.make_callback_wrapper()
	if cCallbackWrapper == nil {
		return nil, nil, errors.New("New async callback failed")
	}

	callbackHolder := unsafe.Pointer(&fn)

	cCallbackWrapper.fn = callbackHolder

	return C.gpointer(unsafe.Pointer(cCallbackWrapper)), callbackHolder, nil
}
