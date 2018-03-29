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

//export goGAsyncReadyCallback
func goGAsyncReadyCallback(result unsafe.Pointer, userData C.gpointer) {
	cCallbackWrapper := (*C.callback_wrapper)(unsafe.Pointer(userData))
	defer C.free_callback_wrapper(cCallbackWrapper) // Free C.make_callback_wrapper()

	fnPtr := (*func(*C.GAsyncResult))(cCallbackWrapper.fn)
	cResult := (*C.GAsyncResult)(result)

	(*fnPtr)(cResult)
}

func newGAsyncReadyCallback(fnPtr *func(*C.GAsyncResult)) (C.GAsyncReadyCallback, C.gpointer, error) {
	if fnPtr == nil {
		return nil, nil, errors.New("Callback can not be nil")
	}

	cCallbackWrapper := C.make_callback_wrapper()
	if cCallbackWrapper == nil {
		return nil, nil, errors.New("New async callback failed")
	}

	cCallbackWrapper.fn = unsafe.Pointer(fnPtr)

	return C.GAsyncReadyCallback(C.my_g_async_ready_callback), C.gpointer(unsafe.Pointer(cCallbackWrapper)), nil
}
