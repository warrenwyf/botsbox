package webkit

/*
#cgo pkg-config: webkit2gtk-4.0

#include <stdlib.h>
#include <webkit2/webkit2.h>

static WebKitWebView* gtk_to_webkit_web_view(GtkWidget *widget)
{
	return WEBKIT_WEB_VIEW(widget);
}
*/
import "C"

import (
	"bytes"
	"unsafe"

	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"

	"../../util"
)

const (
	LoadEvent_Started = iota
	LoadEvent_Redirected
	LoadEvent_Committed
	LoadEvent_Finished
)

const (
	bufSize = 4096
)

//http://webkitgtk.org/reference/webkit2gtk/stable/WebKitWebView.html
type WebView struct {
	cWebView   *C.WebKitWebView
	cGtkWidget *C.GtkWidget
	window     *gtk.Widget

	// cproxy *C.WebKitNetworkProxySettings
}

func NewWebView() *WebView {
	cGtkWidget := C.webkit_web_view_new()
	cWebView := C.gtk_to_webkit_web_view(cGtkWidget)
	gobj := &glib.Object{glib.ToGObject(unsafe.Pointer(cGtkWidget))}

	return &WebView{
		cWebView:   cWebView,
		cGtkWidget: cGtkWidget,
		window: &gtk.Widget{ // Offscreen window
			glib.InitiallyUnowned{gobj},
		},

		// cproxy: nil,
	}
}

// Notice: only call in gtk thread
func (self *WebView) GetWindow() *gtk.Widget {
	return self.window
}

func (self *WebView) GetSettings() *Settings {
	return newSettings(C.webkit_web_view_get_settings(self.cWebView))
}

// func (self *WebView) SetProxy(proxy string) {
// 	if self.cproxy != nil {
// 		C.webkit_network_proxy_settings_free(self.cproxy)
// 	}

// 	cstr := C.CString(proxy)
// 	defer C.free(unsafe.Pointer(cstr))
// 	self.cproxy = C.webkit_network_proxy_settings_new((*C.gchar)(cstr), nil)

// 	ctx := C.webkit_web_view_get_context(self.cWebView)
// 	C.webkit_web_context_set_network_proxy_settings(ctx, C.PROXY_CUSTOM, self.cproxy)
// }

func (self *WebView) LoadURI(uri string) {
	cstr := C.CString(uri)
	defer C.free(unsafe.Pointer(cstr))

	C.webkit_web_view_load_uri(self.cWebView, (*C.gchar)(cstr))
}

func (self *WebView) GetTitle() string {
	return C.GoString((*C.char)(C.webkit_web_view_get_title(self.cWebView)))
}

func (self *WebView) ExportMHtml(c chan<- []byte) unsafe.Pointer {
	if c == nil {
		return nil
	}

	defer func() {
		recover()
	}()

	var asyncCallback = func(result *C.GAsyncResult) {
		if result == nil {
			c <- nil
			return
		}

		// var saveErr *C.GError
		stream := C.webkit_web_view_save_finish(self.cWebView, result, nil /*&saveErr*/)
		if stream == nil {
			// C.g_error_free(saveErr)
			c <- nil
			return

		}

		buf := util.BytesBufferPool.Get().(*bytes.Buffer)
		defer util.BytesBufferPool.Put(buf)

		buf.Reset()

		for {
			// var readErr *C.GError
			cBytes := C.g_input_stream_read_bytes(stream, bufSize, nil, nil /*&readErr*/)
			if cBytes == nil {
				// C.g_error_free(readErr)
				break
			}

			cSize := C.g_bytes_get_size(cBytes)
			if cSize <= 0 {
				break
			}

			p := C.g_bytes_unref_to_data(cBytes, &cSize)
			b := C.GoBytes(unsafe.Pointer(p), C.int(cSize))
			C.g_free(p)

			buf.Write(b)
		}

		C.g_input_stream_close(stream, nil, nil)

		c <- buf.Bytes()
	}

	callbackWrapper, callbackHolder, err := makeCallbackCgo(asyncCallback)
	if err != nil {
		close(c)
		return nil
	}

	C.webkit_web_view_save(self.cWebView, C.WEBKIT_SAVE_MODE_MHTML, nil, asyncReadyCallback, callbackWrapper)

	return callbackHolder
}

// Notice: only call in gtk thread
func (self *WebView) Destroy() {
	// if self.cproxy != nil {
	// 	C.webkit_network_proxy_settings_free(self.cproxy)
	// 	self.cproxy = nil
	// }

	C.webkit_web_view_try_close(self.cWebView)
	C.gtk_widget_destroy(self.cGtkWidget)

	self.window = nil
	self.cGtkWidget = nil
	self.cWebView = nil
}
