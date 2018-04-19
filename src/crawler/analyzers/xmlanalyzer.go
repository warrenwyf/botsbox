package analyzers

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"path"
	"strings"

	"github.com/beevik/etree"
	"golang.org/x/net/html/charset"

	"../../common/util"
	"../rule"
	"../sink"
	"../target"
)

type XmlAnalyzer struct {
	rule *rule.Rule

	doc        *etree.Document
	baseTarget *target.Target

	raw []byte
}

func NewXmlAnalyzer(rule *rule.Rule) *XmlAnalyzer {
	return &XmlAnalyzer{
		rule: rule,
	}
}

func (self *XmlAnalyzer) ParseBytes(b []byte, contentType string, baseTarget *target.Target) (*Result, error) {
	if b == nil {
		return nil, errors.New("Can not parse nil")
	}

	r, errCharset := charset.NewReader(bytes.NewReader(b), contentType) // Determine charset
	if errCharset != nil {
		return nil, errCharset
	}

	return self.readAndParse(r, contentType, baseTarget)
}

func (self *XmlAnalyzer) readAndParse(reader io.Reader, contentType string, baseTarget *target.Target) (*Result, error) {
	doc := etree.NewDocument()
	doc.ReadSettings.Permissive = true

	b := cleanXmlCharacter(reader)
	self.raw = b

	err := doc.ReadFromBytes(b)
	if err != nil {
		return nil, err
	}

	self.doc = doc
	self.baseTarget = baseTarget

	return self.parse()
}

func (self *XmlAnalyzer) parse() (*Result, error) {
	if self.doc == nil || self.baseTarget == nil {
		return nil, errors.New("Can not parse with nil")
	}

	result := &Result{
		Targets:   []*target.Target{},
		SinkPacks: []*sink.SinkPack{},
	}

	// Get Mtag value or use md5(raw content)
	result.Mtag = self.extractXmlValue(self.doc.Root(), self.baseTarget.Mtag)
	if len(result.Mtag) == 0 {
		raw := self.extractXmlValue(self.doc.Root(), "$raw")
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
			for _, s := range self.doc.FindElements(selector) {
				t := target.NewTargetWithTemplate(targetTemplate)
				if t != nil {
					t.Url = actOnXmlUrl(entry.Url, s, self.baseTarget.Url)
					result.Targets = append(result.Targets, t)
				}
			}

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

func (self *XmlAnalyzer) parseObjectOutput(output *rule.ObjectOutput) ([]*target.Target, *sink.SinkPack) {
	name := output.Name
	dataTpl := output.Data

	targets := []*target.Target{}

	id := output.Id
	data := map[string]interface{}{}

	for k, pipeline := range dataTpl {
		v := self.extractXmlValue(self.doc.Root(), pipeline)
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

	value := self.extractXmlValue(self.doc.Root(), id)
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

func (self *XmlAnalyzer) parseListOutput(output *rule.ListOutput) ([]*target.Target, []*sink.SinkPack) {
	name := output.Name
	selector := output.Selector
	dataTpl := output.Data

	targets := []*target.Target{}
	packs := []*sink.SinkPack{}

	for _, s := range self.doc.FindElements(selector) {
		id := output.Id
		data := map[string]interface{}{}

		for k, pipeline := range dataTpl {
			v := self.extractXmlValue(s, pipeline)
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

		value := self.extractXmlValue(self.doc.Root(), id)
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
	}

	return targets, packs
}

func actOnXmlSelection(element *etree.Element, action string) string {
	if element == nil {
		return ""
	}

	if action == "text" { // $text
		return element.Text()

	} else if strings.HasPrefix(action, "attr[") && strings.HasSuffix(action, "]") { // $attr[href]
		attrName := strings.TrimSuffix(strings.TrimPrefix(action, "attr["), "]")
		attr := element.SelectAttr(attrName)
		if attr != nil {
			return attr.Value
		}

	}

	return ""
}

func actOnXmlUrl(u string, element *etree.Element, parentUrl string) string {
	str := strings.TrimSpace(u)
	if strings.HasPrefix(str, "$") && element != nil {
		action := strings.TrimPrefix(str, "$")
		str = actOnXmlSelection(element, action)
	}

	return relUrlToAbs(str, parentUrl)
}

/**
 * $raw
 * $[selector].$text
 * $[selector].$attr[href]
 */
func (self *XmlAnalyzer) extractXmlValue(element *etree.Element, pipeline string) string {
	if len(pipeline) == 0 {
		return ""
	}

	if pipeline == "$raw" {
		if element == self.doc.Root() && self.raw != nil {
			return string(self.raw)
		} else {
			return xmlElementToString(element)
		}

	} else {
		selectorStr := regAction.FindString(pipeline)
		selector := strings.TrimSuffix(strings.TrimPrefix(selectorStr, "$["), "]")
		action := strings.TrimPrefix(strings.TrimPrefix(pipeline, selectorStr), ".$")
		if len(action) == 0 {
			return ""
		}

		if len(selector) == 0 {
			return actOnXmlSelection(element, action)
		} else {
			s := element.FindElement(selector)
			return actOnXmlSelection(s, action)
		}
	}

	return ""
}

func xmlElementToString(element *etree.Element) string {
	if element == nil {
		return ""
	}

	parent := element.Parent()

	doc := etree.NewDocument()
	doc.SetRoot(element)
	str, _ := doc.WriteToString()
	doc.RemoveChild(element)

	if parent != nil {
		parent.AddChild(element)
	}

	return str
}

func cleanXmlCharacter(reader io.Reader) []byte {
	buf := util.BytesBufferPool.Get().(*bytes.Buffer)
	defer util.BytesBufferPool.Put(buf)

	buf.Reset()

	w := bufio.NewWriter(buf)

	br := bufio.NewReader(reader)
	for {
		r, _, err := br.ReadRune()
		if err == io.EOF {
			break
		}

		if !isIllegalXmlCharacterRange(r) {
			continue
		}

		w.WriteRune(r)
	}

	w.Flush()

	return buf.Bytes()
}

func isIllegalXmlCharacterRange(r rune) (inrange bool) {
	return r == 0x09 ||
		r == 0x0A ||
		r == 0x0D ||
		r >= 0x20 && r <= 0xDF77 ||
		r >= 0xE000 && r <= 0xFFFD ||
		r >= 0x10000 && r <= 0x10FFFF
}
