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

type Result struct {
	Targets   []*target.Target
	SinkPacks []*sink.SinkPack
}

func actOnSelection(s *goquery.Selection, action string) string {
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
	if strings.HasPrefix(str, "$") {
		action := strings.TrimPrefix(str, "$")
		str = actOnSelection(s, action)
	}

	/**
	* Convert to absolute URL
	 */

	test, errTest := url.Parse(str)
	if errTest != nil {
		return str
	}

	// Return absolute URL directly
	if test.IsAbs() {
		return str
	}

	// Parse parent URL
	base, errBase := url.Parse(parentUrl)
	if errBase != nil {
		return str
	}

	// Protocol-relative URL
	if strings.HasPrefix(str, "//") {
		return fmt.Sprintf(`%s:%s`, base.Scheme, str)
	}

	// Relative URL
	rel, errRel := base.Parse(str)
	if errRel != nil {
		return str
	}

	return rel.String()
}

func extractOutputValue(doc *goquery.Document, pipeline string) interface{} {
	if pipeline == "$raw" {
		html, err := doc.Html()
		if err == nil {
			return html
		}

	} else if pipeline == "$title" {
		return doc.Find("title").Text()

	} else {
		reg := regexp.MustCompile(`^\$\[*.+\]`)
		selectorStr := reg.FindString(pipeline)
		selector := strings.TrimSuffix(strings.TrimPrefix(selectorStr, "$["), "]")
		action := strings.TrimPrefix(strings.TrimPrefix(pipeline, selectorStr), ".$")
		if len(selector) == 0 || len(action) == 0 {
			return nil
		}

		values := []string{}
		doc.Find(selector).Each(func(i int, s *goquery.Selection) {
			value := actOnSelection(s, action)
			values = append(values, value)
		})

		if len(values) == 0 {
			return nil
		} else if len(values) == 1 {
			return values[0]
		} else {
			return values
		}
	}

	return nil
}
