package analyzers

import (
	"testing"
)

func Test_actOnUrl(t *testing.T) {
	var url string

	url = actOnUrl("http://xyz.com/foo/bar", nil, "https://abc.com/foo/bar.html")
	if url != "http://xyz.com/foo/bar" {
		t.Fatalf(`Wrong URL result: "%s"`, url)
	}

	url = actOnUrl("//xyz.com/foo/bar", nil, "https://abc.com/foo/bar.html")
	if url != "https://xyz.com/foo/bar" {
		t.Fatalf(`Wrong URL result: "%s"`, url)
	}

	url = actOnUrl("./bar", nil, "https://abc.com/foo/bar.html")
	if url != "https://abc.com/foo/bar" {
		t.Fatalf(`Wrong URL result: "%s"`, url)
	}

	url = actOnUrl("/bar", nil, "https://abc.com/foo/bar.html")
	if url != "https://abc.com/bar" {
		t.Fatalf(`Wrong URL result: "%s"`, url)
	}
}
