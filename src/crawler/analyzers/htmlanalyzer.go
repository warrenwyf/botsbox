package analyzers

import (
	"bytes"
	"path"
	"strings"

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

func (self *HtmlAnalyzer) ParseBrowser(b []byte, baseTarget *target.Target) (*Result, error) {
	doc, errParse := goquery.NewDocumentFromReader(bytes.NewReader(b))
	if errParse != nil { // Parse error
		return nil, errParse
	}

	return self.parseDoc(doc.Selection, baseTarget)
}

func (self *HtmlAnalyzer) ParseBytes(b []byte, baseTarget *target.Target) (*Result, error) {
	doc, errParse := goquery.NewDocumentFromReader(bytes.NewReader(b))
	if errParse != nil { // Parse error
		return nil, errParse
	}

	return self.parseDoc(doc.Selection, baseTarget)
}

func (self *HtmlAnalyzer) parseDoc(doc *goquery.Selection, baseTarget *target.Target) (*Result, error) {
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
			t := target.NewTargetWithJson(&targetTemplateElem)
			if t != nil {
				t.Url = actOnUrl(entry.Url, s, baseTarget.Url)
				result.Targets = append(result.Targets, t)
			}
		})
	}

	// Analyze object outputs
	for _, objectOutput := range baseTarget.ObjectOutputs {
		name := objectOutput.Name
		if len(name) == 0 {
			continue
		}

		targets, pack := self.parseObjectOutput(doc, objectOutput, baseTarget)

		result.Targets = append(result.Targets, targets...)
		result.SinkPacks = append(result.SinkPacks, pack)
	}

	// Analyze list outputs
	for _, listOutput := range baseTarget.ListOutputs {
		name := listOutput.Name
		if len(name) == 0 {
			continue
		}

		targets, packs := self.parseListOutput(doc, listOutput, baseTarget)

		result.Targets = append(result.Targets, targets...)
		result.SinkPacks = append(result.SinkPacks, packs...)
	}

	return result, nil
}

func (self *HtmlAnalyzer) parseObjectOutput(doc *goquery.Selection, output *target.ObjectOutput, baseTarget *target.Target) ([]*target.Target, *sink.SinkPack) {
	name := output.Name
	dataTpl := output.Data

	targets := []*target.Target{}
	data := map[string]interface{}{}

	for k, pipeline := range dataTpl {
		v := extractHtmlValue(doc, pipeline)
		if len(v) == 0 {
			continue
		}

		// k may have extension
		ext := path.Ext(k) // .xxx
		if len(ext) > 1 {
			contentType := strings.TrimPrefix(ext, ".")

			// Fetch additional file target
			url := relUrlToAbs(v, baseTarget.Url)
			t := newFileTarget(name, url, contentType)
			targets = append(targets, t)

			data[k] = url
		} else {
			data[k] = v
		}
	}

	pack := &sink.SinkPack{
		Name: name,
		Hash: baseTarget.GetResult().Hash,
		Url:  baseTarget.Url,
		Data: data,
	}

	return targets, pack
}

func (self *HtmlAnalyzer) parseListOutput(doc *goquery.Selection, output *target.ListOutput, baseTarget *target.Target) ([]*target.Target, []*sink.SinkPack) {
	name := output.Name
	selector := output.Selector
	dataTpl := output.Data

	targets := []*target.Target{}
	packs := []*sink.SinkPack{}

	doc.Find(selector).Each(func(i int, s *goquery.Selection) {
		data := map[string]interface{}{}

		for k, pipeline := range dataTpl {
			v := extractHtmlValue(s, pipeline)
			if len(v) == 0 {
				continue
			}

			// k may have extension
			ext := path.Ext(k) // .xxx
			if len(ext) > 1 {
				contentType := strings.TrimPrefix(ext, ".")

				// Fetch additional file target
				url := relUrlToAbs(v, baseTarget.Url)
				t := newFileTarget(name, url, contentType)
				targets = append(targets, t)

				data[k] = url
			} else {
				data[k] = v
			}
		}

		pack := &sink.SinkPack{
			Name: name,
			Hash: baseTarget.GetResult().Hash,
			Url:  baseTarget.Url,
			Data: data,
		}

		packs = append(packs, pack)
	})

	return targets, packs
}
