package target

import (
	"testing"

	"../rule"
)

func Test_MakeTargetsWithRule(t *testing.T) {
	ruleContent := `
	{
		"$every": "5m",
		"$entries": [
			{
				"$name": "recommend_api",
				"$url": "https://api-site",
				"$method": "post",
				"$form": {
					"method": "next",
					"params": "{\"limit\":20,\"offset\":$var[offset]}"
				},
				"$contentType": "json",
				"$var": {
					"offset": "$rangeInt[0, 100, 20]"
				}
			}
		],
		"recommend_api": {
			"$outputs": [
				{
					"$name": "zhihu_recommend",
					"$data": {
						"result": "$raw"
					}
				}
			]
		}
	}
	`

	r, err := rule.NewRuleWithContent(ruleContent)
	if err != nil {
		t.Fatalf("NewRuleWithContent error: %v", err)
	}

	entry := r.Entries[0]
	targetTemplate, ok := r.TargetTemplates["recommend_api"]
	if !ok {
		t.Fatal("Missing target template")
	}

	targets := MakeTargetsWithRule(entry, targetTemplate)
	if len(targets) == 0 {
		t.Fatal("MakeTargetsWithRule failed")
	}

}
