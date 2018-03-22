package fetcher

import (
	"bytes"
	"testing"

	"github.com/PuerkitoBio/goquery"
	//"github.com/tidwall/gjson"
)

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
	url := "https://news.baidu.com/"

	f := NewHttpFetcher()
	f.SetUrl(url)

	result, errFetch := f.Fetch()
	if errFetch != nil {
		t.Fatalf(`Fetch "%s" error: %v`, url, errFetch)
	}

	doc, errParse := goquery.NewDocumentFromReader(bytes.NewReader(result.Content.([]byte)))
	if errParse != nil {
		t.Fatalf(`Parse html content error: %v`, errParse)
	}

	elems := doc.Find("#header .logo img")
	if elems.Length() == 0 {
		t.Fatalf(`No logo exists`)
	}

	elems.Each(func(i int, s *goquery.Selection) {
		src, ok := s.Attr("src")
		if ok {
			t.Logf("#logo has src: %v", src)
		} else {
			t.Fatalf(`No src in logo`)
		}
	})

}
