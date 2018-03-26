package analyzers

import (
	"../rule"
	"../sink"
	"../target"
)

type BinaryAnalyzer struct {
	rule *rule.Rule
}

func NewBinaryAnalyzer(rule *rule.Rule) *BinaryAnalyzer {
	return &BinaryAnalyzer{
		rule: rule,
	}
}

func (self *BinaryAnalyzer) Parse(b []byte, baseTarget *target.Target) (*Result, error) {
	result := &Result{}

	baseResult := baseTarget.GetResult()

	// Analyze outputs
	for _, output := range baseTarget.ObjectOutputs {
		// Pack & Send
		name := output.Name
		if len(name) > 0 {
			pack := &sink.SinkPack{
				Name:    name,
				Hash:    baseResult.Hash,
				Url:     baseTarget.Url,
				File:    b,
				FileExt: "." + baseTarget.ContentType,
			}

			result.SinkPacks = append(result.SinkPacks, pack)
		}
	}

	return result, nil
}
