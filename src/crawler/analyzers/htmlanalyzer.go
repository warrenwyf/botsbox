package analyzers

import (
	"bytes"
	"errors"
	"io"
	"path"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html/charset"

	"../../common/util"
	"../rule"
	"../sink"
	"../target"
)

type HtmlAnalyzer struct {
	rule *rule.Rule

	doc        *goquery.Document
	baseTarget *target.Target

	raw []byte
}

func NewHtmlAnalyzer(rule *rule.Rule) *HtmlAnalyzer {
	return &HtmlAnalyzer{
		rule: rule,
	}
}

func (self *HtmlAnalyzer) ParseBytes(b []byte, contentType string, baseTarget *target.Target) (*Result, error) {
	if b == nil {
		return nil, errors.New("Can not parse nil")
	}

	r, errCharset := charset.NewReader(bytes.NewReader(b), contentType) // Determine charset
	if errCharset != nil {
		return nil, errCharset
	}

	return self.readAndParse(r, contentType, baseTarget)
}

func (self *HtmlAnalyzer) readAndParse(reader io.Reader, contentType string, baseTarget *target.Target) (*Result, error) {
	buf := util.BytesBufferPool.Get().(*bytes.Buffer)
	defer util.BytesBufferPool.Put(buf)

	buf.Reset()
	io.TeeReader(reader, buf)

	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		return nil, err
	}

	// After TeeReader read via goquery
	self.raw = buf.Bytes()

	self.doc = doc
	self.baseTarget = baseTarget

	return self.parse()
}

func (self *HtmlAnalyzer) parse() (*Result, error) {
	if self.doc == nil || self.baseTarget == nil {
		return nil, errors.New("Can not parse with nil")
	}

	result := &Result{
		Targets:   []*target.Target{},
		SinkPacks: []*sink.SinkPack{},
	}

	// Get Mtag value or use md5(raw content)
	result.Mtag = self.extractHtmlValue(self.doc.Selection, self.baseTarget.Mtag)
	if len(result.Mtag) == 0 {
		raw := self.extractHtmlValue(self.doc.Selection, "$raw")
		result.Mtag = util.Md5Bytes([]byte(raw))
	}

	// Analyze deeper targets
	for selector, entry := range self.baseTarget.Dive {
		targetTemplate, ok := self.rule.TargetTemplates[entry.Name]
		if !ok {
			continue
		}

		if strings.HasPrefix(selector, "$") { // Virtual selector, means dive directly
			t := target.NewTargetWithTemplate(targetTemplate)
			if t != nil {
				t.Url = relUrlToAbs(entry.Url, self.baseTarget.Url)
				result.Targets = append(result.Targets, t)
			}

		} else {
			self.doc.Find(selector).Each(func(i int, s *goquery.Selection) {
				t := target.NewTargetWithTemplate(targetTemplate)
				if t != nil {
					t.Url = actOnHtmlUrl(entry.Url, s, self.baseTarget.Url)
					result.Targets = append(result.Targets, t)
				}
			})

		}

	}

	// Analyze object outputs
	for _, objectOutput := range self.baseTarget.ObjectOutputs {
		name := objectOutput.Name
		if len(name) == 0 {
			continue
		}

		targets, pack := self.parseObjectOutput(objectOutput)

		result.Targets = append(result.Targets, targets...)
		result.SinkPacks = append(result.SinkPacks, pack)
	}

	// Analyze list outputs
	for _, listOutput := range self.baseTarget.ListOutputs {
		name := listOutput.Name
		if len(name) == 0 {
			continue
		}

		targets, packs := self.parseListOutput(listOutput)

		result.Targets = append(result.Targets, targets...)
		result.SinkPacks = append(result.SinkPacks, packs...)
	}

	return result, nil
}

func (self *HtmlAnalyzer) parseObjectOutput(output *rule.ObjectOutput) ([]*target.Target, *sink.SinkPack) {
	name := output.Name
	dataTpl := output.Data

	targets := []*target.Target{}

	id := output.Id
	data := map[string]interface{}{}

	for k, pipeline := range dataTpl {
		v := self.extractHtmlValue(self.doc.Selection, pipeline)
		if len(v) == 0 {
			continue
		}

		// k may have extension
		ext := path.Ext(k) // .xxx
		if len(ext) > 1 {
			resultType := strings.TrimPrefix(ext, ".")

			// Fetch additional file target
			url := relUrlToAbs(v, self.baseTarget.Url)
			t := newFileTarget(name, url, resultType)
			targets = append(targets, t)

			data[k] = url
		} else {
			data[k] = v
		}
	}

	value := self.extractHtmlValue(self.doc.Selection, id)
	if len(value) > 0 {
		id = value
	} else {
		for varName, varValue := range self.baseTarget.ApplyedVar {
			id = rule.ApplyVarToString(id, varName, varValue)
		}
	}

	var hash string = ""
	if baseTargetResult := self.baseTarget.GetFetchResult(); baseTargetResult != nil {
		hash = baseTargetResult.Hash
	}

	pack := &sink.SinkPack{
		Name: name,

		Id:   id,
		Hash: hash,
		Data: data,
	}

	return targets, pack
}

func (self *HtmlAnalyzer) parseListOutput(output *rule.ListOutput) ([]*target.Target, []*sink.SinkPack) {
	name := output.Name
	selector := output.Selector
	dataTpl := output.Data

	targets := []*target.Target{}
	packs := []*sink.SinkPack{}

	self.doc.Find(selector).Each(func(i int, s *goquery.Selection) {
		id := output.Id
		data := map[string]interface{}{}

		for k, pipeline := range dataTpl {
			v := self.extractHtmlValue(s, pipeline)
			if len(v) == 0 {
				continue
			}

			// k may have extension
			ext := path.Ext(k) // .xxx
			if len(ext) > 1 {
				resultType := strings.TrimPrefix(ext, ".")

				// Fetch additional file target
				url := relUrlToAbs(v, self.baseTarget.Url)
				t := newFileTarget(name, url, resultType)
				targets = append(targets, t)

				data[k] = url
			} else {
				data[k] = v
			}
		}

		value := self.extractHtmlValue(s, id)
		if len(value) > 0 {
			id = value
		} else {
			for varName, varValue := range self.baseTarget.ApplyedVar {
				id = rule.ApplyVarToString(id, varName, varValue)
			}
		}

		var hash string = ""
		if baseTargetResult := self.baseTarget.GetFetchResult(); baseTargetResult != nil {
			hash = baseTargetResult.Hash
		}

		pack := &sink.SinkPack{
			Name: name,

			Id:   id,
			Hash: hash,
			Data: data,
		}

		packs = append(packs, pack)
	})

	return targets, packs
}

func actOnHtmlSelection(element *goquery.Selection, action string) string {
	if element == nil {
		return ""
	}

	if action == "text" { // $text
		return element.Text()

	} else if action == "html" { // $html
		html, err := element.Html()
		if err == nil {
			return html
		}

	} else if strings.HasPrefix(action, "attr[") && strings.HasSuffix(action, "]") { // $attr[href]
		attrName := strings.TrimSuffix(strings.TrimPrefix(action, "attr["), "]")
		attr, attrExist := element.Attr(attrName)
		if attrExist {
			return attr
		}

	}

	return ""
}

func actOnHtmlUrl(u string, element *goquery.Selection, parentUrl string) string {
	str := strings.TrimSpace(u)
	if strings.HasPrefix(str, "$") && element != nil {
		action := strings.TrimPrefix(str, "$")
		str = actOnHtmlSelection(element, action)
	}

	return relUrlToAbs(str, parentUrl)
}

/**
 * $raw
 * $title
 * $[selector].$text
 * $[selector].$html
 * $[selector].$attr[href]
 */
func (self *HtmlAnalyzer) extractHtmlValue(element *goquery.Selection, pipeline string) string {
	if len(pipeline) == 0 {
		return ""
	}

	if pipeline == "$raw" {
		if element == self.doc.Selection && self.raw != nil {
			return string(self.raw)
		} else {
			html, err := element.Html()
			if err == nil {
				return html
			}
		}

	} else if pipeline == "$title" {
		return element.Find("title").Text()

	} else {
		selectorStr := regAction.FindString(pipeline)
		selector := strings.TrimSuffix(strings.TrimPrefix(selectorStr, "$["), "]")
		action := strings.TrimPrefix(strings.TrimPrefix(pipeline, selectorStr), ".$")
		if len(selector) == 0 || len(action) == 0 {
			return ""
		}

		s := element.Find(selector).First()
		return actOnHtmlSelection(s, action)

	}

	return ""
}
