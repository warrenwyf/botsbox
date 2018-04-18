package analyzers

import (
	"errors"
	"path"
	"strings"

	"github.com/tidwall/gjson"

	"../../common/util"
	"../rule"
	"../sink"
	"../target"
)

type JsonAnalyzer struct {
	rule *rule.Rule

	doc        gjson.Result
	baseTarget *target.Target
}

func NewJsonAnalyzer(rule *rule.Rule) *JsonAnalyzer {
	return &JsonAnalyzer{
		rule: rule,
	}
}

func (self *JsonAnalyzer) ParseBytes(json []byte, baseTarget *target.Target) (*Result, error) {
	if json == nil {
		return nil, errors.New("Can not parse with nil")
	}

	self.doc = gjson.ParseBytes(json)
	self.baseTarget = baseTarget

	return self.parse()
}

func (self *JsonAnalyzer) parse() (*Result, error) {
	if !self.doc.Exists() || self.baseTarget == nil {
		return nil, errors.New("Can not parse with nil")
	}

	result := &Result{
		Targets:   []*target.Target{},
		SinkPacks: []*sink.SinkPack{},
	}

	// Get Mtag value or use md5(raw content)
	result.Mtag = self.extractJsonValue(self.baseTarget.Mtag)
	if len(result.Mtag) == 0 {
		raw := self.extractJsonValue("$raw")
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
			self.doc.Get(selector).ForEach(func(k, v gjson.Result) bool {
				t := target.NewTargetWithTemplate(targetTemplate)
				if t != nil {
					t.Url = v.String()
					result.Targets = append(result.Targets, t)
				}

				return true
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

func (self *JsonAnalyzer) parseObjectOutput(output *rule.ObjectOutput) ([]*target.Target, *sink.SinkPack) {
	name := output.Name
	id := output.Id
	dataTpl := output.Data

	targets := []*target.Target{}
	data := map[string]interface{}{}

	for k, pipeline := range dataTpl {
		v := self.extractJsonValue(pipeline)
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

	value := self.extractJsonValue(id)
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

func (self *JsonAnalyzer) parseListOutput(output *rule.ListOutput) ([]*target.Target, []*sink.SinkPack) {
	name := output.Name
	selector := output.Selector
	id := output.Id
	dataTpl := output.Data

	targets := []*target.Target{}
	packs := []*sink.SinkPack{}

	self.doc.Get(selector).ForEach(func(k, v gjson.Result) bool {
		data := map[string]interface{}{}

		for k, pipeline := range dataTpl {
			v := self.extractJsonValue(pipeline)
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

		value := self.extractJsonValue(id)
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

		return true
	})

	return targets, packs
}

/**
 * $raw
 * $[selector]
 */
func (self *JsonAnalyzer) extractJsonValue(pipeline string) string {
	if len(pipeline) == 0 {
		return ""
	}

	if pipeline == "$raw" {
		return self.doc.Raw

	} else {
		selectorStr := regAction.FindString(pipeline)
		selector := strings.TrimSuffix(strings.TrimPrefix(selectorStr, "$["), "]")
		if len(selector) == 0 {
			return ""
		}

		return self.doc.Get(selector).String()
	}

	return ""
}
