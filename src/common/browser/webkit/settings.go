package webkit

/*
#include <webkit2/webkit2.h>
*/
import "C"

import (
	"unsafe"
)

func gbool(b bool) C.gboolean {
	if b {
		return C.gboolean(1)
	}
	return C.gboolean(0)
}

type Settings struct {
	cs *C.WebKitSettings
}

func newSettings(cs *C.WebKitSettings) *Settings {
	return &Settings{
		cs: cs,
	}
}

func (self *Settings) SetAutoLoadImages(v bool) {
	C.webkit_settings_set_auto_load_images(self.cs, gbool(v))
}

func (self *Settings) SetUserAgent(v string) {
	cstr := C.CString(v)
	defer C.free(unsafe.Pointer(cstr))

	C.webkit_settings_set_user_agent(self.cs, (*C.gchar)(cstr))
}

func (self *Settings) SetDefaultCharset(v string) {
	cstr := C.CString(v)
	defer C.free(unsafe.Pointer(cstr))

	C.webkit_settings_set_default_charset(self.cs, (*C.gchar)(cstr))
}
