package target

import (
	"../rule"
)

func MakeTargetsWithRule(entry *rule.Entry, targetTemplate *rule.TargetTemplate) []*Target {
	// Parameterized targets
	if len(entry.Var) > 0 {
		targets := []*Target{}

		minCount := 0
		for _, varValues := range entry.Var {
			count := len(varValues)
			if count < minCount || minCount == 0 {
				minCount = count
			}
		}

		for i := 0; i < minCount; i++ {
			t := newTargetWithRule(entry, targetTemplate)
			if t == nil {
				continue
			}

			for varName, varValues := range entry.Var {
				t.ApplyedVar[varName] = varValues[i]

				t.Url = rule.ApplyVarToString(t.Url, varName, varValues[i])
				t.Query = rule.ApplyVarToMap(t.Query, varName, varValues[i])
				t.Form = rule.ApplyVarToMap(t.Form, varName, varValues[i])

				targets = append(targets, t)
			}
		}

		return targets
	}

	return []*Target{newTargetWithRule(entry, targetTemplate)}
}

func newTargetWithRule(entry *rule.Entry, targetTemplate *rule.TargetTemplate) *Target {
	t := NewTargetWithTemplate(targetTemplate)

	t.Url = entry.Url
	t.Method = entry.Method
	t.Header = entry.Header
	t.Query = entry.Query
	t.Form = entry.Form
	t.ResultType = entry.ResultType

	return t
}
