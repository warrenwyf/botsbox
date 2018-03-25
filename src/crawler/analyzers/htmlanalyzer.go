package analyzers

import (
	"bytes"

	"github.com/PuerkitoBio/goquery"

	"../rule"
	"../sink"
	"../target"
)

type HtmlAnalyzer struct {
	rule *rule.Rule
}

func NewHtmlAnalyzer(rule *rule.Rule) *HtmlAnalyzer {
	return &HtmlAnalyzer{
		rule: rule,
	}
}

func (self *HtmlAnalyzer) ParseBytes(b []byte, baseTarget *target.Target) (*Result, error) {
	doc, errParse := goquery.NewDocumentFromReader(bytes.NewReader(b))
	if errParse != nil { // Parse error
		return nil, errParse
	}

	result := &Result{
		Targets:   []*target.Target{},
		SinkPacks: []*sink.SinkPack{},
	}

	// Analyze deeper targets
	for selector, entry := range baseTarget.Dive {
		targetTemplateElem := self.rule.GetTargetTemplate(entry.Name)
		if !targetTemplateElem.Exists() {
			continue
		}

		doc.Find(selector).Each(func(i int, s *goquery.Selection) {
			t := target.NewTargetWithJson(targetTemplateElem)
			if t != nil {
				t.Url = actOnUrl(entry.Url, s, baseTarget.Url)
				result.Targets = append(result.Targets, t)
			}
		})
	}

	// Analyze outputs
	for _, output := range baseTarget.Outputs {
		data := map[string]interface{}{}

		dataTpl := output.Data
		for k, action := range dataTpl {
			v := extractOutputValue(doc, action)
			if v != nil {
				data[k] = v
			}
		}

		// Pack & Send
		name := output.Name
		if len(name) > 0 {
			pack := &sink.SinkPack{
				Name: name,
				Url:  baseTarget.Url,
				Data: data,
			}

			result.SinkPacks = append(result.SinkPacks, pack)
		}
	}

	return result, nil
}
