package analyzers

import (
	"../../common/util"
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

	result.Mtag = util.Md5Bytes(b)

	baseResult := baseTarget.GetResult()

	// Analyze outputs
	for _, output := range baseTarget.ObjectOutputs {
		// Pack & Send
		name := output.Name
		if len(name) > 0 {
			pack := &sink.SinkPack{
				Name:    name,
				Id:      baseTarget.Url,
				Hash:    baseResult.Hash,
				File:    b,
				FileExt: "." + baseTarget.ResultType,
			}

			result.SinkPacks = append(result.SinkPacks, pack)
		}
	}

	return result, nil
}
