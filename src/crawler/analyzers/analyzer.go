package analyzers

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"

	"../sink"
	"../target"
)

var (
	regAction = regexp.MustCompile(`^\$\[*.+\]`)
)

type Result struct {
	Targets   []*target.Target
	SinkPacks []*sink.SinkPack
}

func actOnSelection(s *goquery.Selection, action string) string {
	if s == nil {
		return ""
	}

	if action == "text" { // $text
		return s.Text()

	} else if action == "html" { // $html
		html, err := s.Html()
		if err == nil {
			return html
		}

	} else if strings.HasPrefix(action, "attr[") && strings.HasSuffix(action, "]") { // $attr[href]
		attrName := strings.TrimSuffix(strings.TrimPrefix(action, "attr["), "]")
		attr, attrExist := s.Attr(attrName)
		if attrExist {
			return attr
		}

	}

	return ""
}

func actOnUrl(u string, s *goquery.Selection, parentUrl string) string {
	str := strings.TrimSpace(u)
	if strings.HasPrefix(str, "$") && s != nil {
		action := strings.TrimPrefix(str, "$")
		str = actOnSelection(s, action)
	}

	return relUrlToAbs(str, parentUrl)
}

func relUrlToAbs(relUrl string, parentUrl string) string {
	test, errTest := url.Parse(relUrl)
	if errTest != nil {
		return relUrl
	}

	// Return absolute URL directly
	if test.IsAbs() {
		return relUrl
	}

	// Parse parent URL
	base, errBase := url.Parse(parentUrl)
	if errBase != nil {
		return relUrl
	}

	// Protocol-relative URL
	if strings.HasPrefix(relUrl, "//") {
		return fmt.Sprintf(`%s:%s`, base.Scheme, relUrl)
	}

	// Absolute URL
	abs, errRel := base.Parse(relUrl)
	if errRel != nil {
		return relUrl
	}

	return abs.String()
}

/**
 * $raw
 * $title
 * $[selector].$text
 * $[selector].$html
 * $[selector].$attr[href]
 */
func extractHtmlValue(doc *goquery.Document, pipeline string) string {
	if pipeline == "$raw" {
		html, err := doc.Html()
		if err == nil {
			return html
		}

	} else if pipeline == "$title" {
		return doc.Find("title").Text()

	} else {
		selectorStr := regAction.FindString(pipeline)
		selector := strings.TrimSuffix(strings.TrimPrefix(selectorStr, "$["), "]")
		action := strings.TrimPrefix(strings.TrimPrefix(pipeline, selectorStr), ".$")
		if len(selector) == 0 || len(action) == 0 {
			return ""
		}

		s := doc.Find(selector).First()
		return actOnSelection(s, action)

	}

	return ""
}

func extractHtmlElementValue(element *goquery.Selection, pipeline string) string {
	if pipeline == "$raw" {
		html, err := element.Html()
		if err == nil {
			return html
		}

	} else {
		selectorStr := regAction.FindString(pipeline)
		selector := strings.TrimSuffix(strings.TrimPrefix(selectorStr, "$["), "]")
		action := strings.TrimPrefix(strings.TrimPrefix(pipeline, selectorStr), ".$")
		if len(selector) == 0 || len(action) == 0 {
			return ""
		}

		s := element.Find(selector).First()
		return actOnSelection(s, action)

	}

	return ""
}

func newFileTarget(dir string, url string, contentType string) *target.Target {
	t := target.NewTarget()
	t.Url = url
	t.ContentType = contentType

	output := target.NewObjectOutput()
	output.Name = dir

	t.ObjectOutputs = []*target.ObjectOutput{output}

	return t
}
