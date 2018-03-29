package webkit

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gotk3/gotk3/gtk"
)

var (
	server  *httptest.Server
	webView *WebView
)

func setup() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(res http.ResponseWriter, req *http.Request) {
		buf := bytes.NewBufferString(req.UserAgent())
		res.Write(buf.Bytes())
	})

	server = httptest.NewServer(mux)

	gtk.Init(nil)
}

func teardown() {
	server.Close()
}

func TestMain(m *testing.M) {
	setup()

	m.Run()

	teardown()
}

func Test_NewWebView(t *testing.T) {
	webView = NewWebView()
	if webView == nil {
		t.Fatal("New WebView error")
	}
}

// func Test_SetProxy(t *testing.T) {
// 	webView.SetProxy("socks://127.0.0.1:1080")
// }

func Test_LoadURI(t *testing.T) {
	webView.LoadURI(server.URL)
}

func Test_GetTitle(t *testing.T) {
	webView.GetTitle()
}

func Test_ExportMHtml(t *testing.T) {
	webView.ExportMHtml(func(bytes []byte) {
		t.Log(bytes)
	})
}

func Test_Destroy(t *testing.T) {
	webView.Destroy()
}

func Test_DestroyServer(t *testing.T) {
	server.Close()
}
