package rule

import (
	"testing"
)

func Test_ListOutput(t *testing.T) {
	ruleContent := `
	{
		"$timeout": "1d",
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

	_, ok := rule.TargetTemplates["index_page"]
	if !ok {
		t.Fatalf("Missing target template")
	}
}

func Test_ParameterizedTarget(t *testing.T) {
	ruleContent := `
	{
		"$every": "5m",
		"$entries": [
			{
				"$name": "recommend_api",
				"$url": "https://www.zhihu.com/node/ExploreRecommendListV2",
				"$method": "post",
				"$form": {
					"method": "next",
					"params": "{\"limit\":20,\"offset\":$var[offset]}"
				},
				"$contentType": "json",
				"$var": {
					"offset": "$[0, 20, 40, 60, 80, 100]",
					"offset": "$rangeInt[0, 100, 20]"
				}
			}
		],
		"recommend_api": {
			"$outputs": [
				{
					"$name": "zhihu_recommend",
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

	r, err := NewRuleWithContent(ruleContent)
	if err != nil {
		t.Fatalf("NewRuleWithContent error: %v", err)
	}

	_, ok := r.TargetTemplates["recommend_api"]
	if !ok {
		t.Fatalf("Missing target template")
	}
}
