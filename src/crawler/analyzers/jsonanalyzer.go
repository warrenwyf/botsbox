package analyzers

import (
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
}

func NewJsonAnalyzer(rule *rule.Rule) *JsonAnalyzer {
	return &JsonAnalyzer{
		rule: rule,
	}
}

func (self *JsonAnalyzer) ParseBytes(json []byte, baseTarget *target.Target) (*Result, error) {
	result := &Result{
		Targets:   []*target.Target{},
		SinkPacks: []*sink.SinkPack{},
	}

	// Get Mtag value or use md5(raw content)
	result.Mtag = extractJsonValue(json, baseTarget.Mtag)
	if len(result.Mtag) == 0 {
		raw := extractJsonValue(json, "$raw")
		result.Mtag = util.Md5Bytes([]byte(raw))
	}

	// Analyze deeper targets
	for selector, entry := range baseTarget.Dive {
		targetTemplate, ok := self.rule.TargetTemplates[entry.Name]
		if !ok {
			continue
		}

		gjson.GetBytes(json, selector).ForEach(func(k, v gjson.Result) bool {
			t := target.NewTargetWithTemplate(targetTemplate)
			if t != nil {
				t.Url = v.String()
				result.Targets = append(result.Targets, t)
			}

			return true
		})
	}

	// Analyze object outputs
	for _, objectOutput := range baseTarget.ObjectOutputs {
		name := objectOutput.Name
		if len(name) == 0 {
			continue
		}

		targets, pack := self.parseObjectOutput(json, objectOutput, baseTarget)

		result.Targets = append(result.Targets, targets...)
		result.SinkPacks = append(result.SinkPacks, pack)
	}

	// Analyze list outputs
	for _, listOutput := range baseTarget.ListOutputs {
		name := listOutput.Name
		if len(name) == 0 {
			continue
		}

		targets, packs := self.parseListOutput(json, listOutput, baseTarget)

		result.Targets = append(result.Targets, targets...)
		result.SinkPacks = append(result.SinkPacks, packs...)
	}

	return result, nil
}

func (self *JsonAnalyzer) parseObjectOutput(json []byte, output *rule.ObjectOutput, baseTarget *target.Target) ([]*target.Target, *sink.SinkPack) {
	name := output.Name
	id := output.Id
	dataTpl := output.Data

	targets := []*target.Target{}
	data := map[string]interface{}{}

	for k, pipeline := range dataTpl {
		v := extractJsonValue(json, pipeline)
		if len(v) == 0 {
			continue
		}

		// k may have extension
		ext := path.Ext(k) // .xxx
		if len(ext) > 1 {
			resultType := strings.TrimPrefix(ext, ".")

			// Fetch additional file target
			url := relUrlToAbs(v, baseTarget.Url)
			t := newFileTarget(name, url, resultType)
			targets = append(targets, t)

			data[k] = url
		} else {
			data[k] = v
		}
	}

	value := extractJsonValue(json, id)
	if len(value) > 0 {
		id = value
	} else {
		for varName, varValue := range baseTarget.ApplyedVar {
			id = rule.ApplyVarToString(id, varName, varValue)
		}
	}

	pack := &sink.SinkPack{
		Name: name,

		Id:   id,
		Hash: baseTarget.GetResult().Hash,
		Data: data,
	}

	return targets, pack
}

func (self *JsonAnalyzer) parseListOutput(json []byte, output *rule.ListOutput, baseTarget *target.Target) ([]*target.Target, []*sink.SinkPack) {
	name := output.Name
	selector := output.Selector
	id := output.Id
	dataTpl := output.Data

	targets := []*target.Target{}
	packs := []*sink.SinkPack{}

	gjson.GetBytes(json, selector).ForEach(func(k, v gjson.Result) bool {
		data := map[string]interface{}{}

		for k, pipeline := range dataTpl {
			v := extractJsonValue(json, pipeline)
			if len(v) == 0 {
				continue
			}

			// k may have extension
			ext := path.Ext(k) // .xxx
			if len(ext) > 1 {
				resultType := strings.TrimPrefix(ext, ".")

				// Fetch additional file target
				url := relUrlToAbs(v, baseTarget.Url)
				t := newFileTarget(name, url, resultType)
				targets = append(targets, t)

				data[k] = url
			} else {
				data[k] = v
			}
		}

		value := extractJsonValue(json, id)
		if len(value) > 0 {
			id = value
		} else {
			for varName, varValue := range baseTarget.ApplyedVar {
				id = rule.ApplyVarToString(id, varName, varValue)
			}
		}

		pack := &sink.SinkPack{
			Name: name,

			Id:   id,
			Hash: baseTarget.GetResult().Hash,
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
func extractJsonValue(json []byte, pipeline string) string {
	if len(pipeline) == 0 {
		return ""
	}

	if pipeline == "$raw" {
		return string(json)

	} else {
		selectorStr := regAction.FindString(pipeline)
		selector := strings.TrimSuffix(strings.TrimPrefix(selectorStr, "$["), "]")
		if len(selector) == 0 {
			return ""
		}

		return gjson.GetBytes(json, selector).String()

	}

	return ""
}
