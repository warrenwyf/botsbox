package fetchers

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

var (
	server *httptest.Server
)

func Test_SetupServer(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(res http.ResponseWriter, req *http.Request) {
		defer req.Body.Close()

		if strings.ToUpper(req.Method) == "GET" {
			res.Write([]byte("get ok"))
		} else if strings.ToUpper(req.Method) == "POST" {
			b, _ := ioutil.ReadAll(req.Body)
			res.Write(b)
		}
	})

	server = httptest.NewServer(mux)
}

func Test_HttpFetcher_Hash(t *testing.T) {
	f1 := NewHttpFetcher()
	f1.SetUrl("https://one-site")

	f2 := NewHttpFetcher()
	f2.SetUrl("https://other-site")

	if f1.Hash() == f2.Hash() {
		t.Fatalf(`Hash error`)
	}
}

func Test_HttpFetcher_Get(t *testing.T) {
	f := NewHttpFetcher()
	f.SetUrl(server.URL)

	result, errFetch := f.Fetch()
	if errFetch != nil {
		t.Fatalf(`Get error: %v`, errFetch)
	}

	if string(result.Content.([]byte)) != "get ok" {
		t.Fatal("Get result wrong")
	}
}

func Test_HttpFetcher_Post(t *testing.T) {
	data := map[string]string{
		"foo": "bar",
	}

	f := NewHttpFetcher()
	f.SetUrl(server.URL)
	f.SetMethod("POST")
	f.SetForm(data)

	result, errFetch := f.Fetch()
	if errFetch != nil {
		t.Fatalf(`Post error: %v`, errFetch)
	}

	if string(result.Content.([]byte)) != "foo=bar" {
		t.Fatal("Post result wrong")
	}
}

func Test_DestroyServer(t *testing.T) {
	server.Close()
}
