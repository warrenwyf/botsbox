package rule

import (
	"testing"
)

func Test_ListOutput(t *testing.T) {
	ruleContent := `
	{
		"$every": "5m",
		"$entries": [
			{
				"$name": "index_page",
				"$url": "https://s.taobao.com/list?q=衬衫"
			}
		],
		"index_page": {
			"$outputs": [
				{
					"$name": "taobao_female_shirt",
					"$each": ".m-itemlist .item",
					"$data": {
						"title": "$[.title a].$text",
						"price": "$[.price strong].$text",
						"deal": "$[.deal-cnt].$text",
						"shopname": "$[.shop .shopname span:last-child].$text",
						"location": "$[.shop .location].$text",
						"cover.webp": "$[.pic img].$attr[src]"
					}
				}
			]
		}
	}
	`

	rule, err := NewRuleWithContent(ruleContent)
	if err != nil {
		t.Fatalf("NewRuleWithContent error: %v", err)
	}

	if !rule.GetTargetTemplate("index_page").Exists() {
		t.Fatalf("Missing target template")
	}
}
