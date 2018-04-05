package job

import (
	"testing"
)

var job *Job

func Test_NewJob(t *testing.T) {
	rule := `
	{
		"$every": "5m",
		"$startDay": "w0",
		"$startDayTime": "03:00:00",
		"$entries": [
			{
				"$name": "index_page",
				"$url": "https://news.baidu.com"
			}
		],
		"index_page": {
			"$dive": {
				".hotnews li a": {
					"$name": "hotnews_page",
					"$url": "$attr[href]"
				}
			}
		},
		"hotnews_page": {
			"$priority": 5,
			"$retry": 5,
			"$retryWait": "30s",
			"$outputs": [
				{
					"$name": "baidu_hotnews",
					"$data": {
						"title": "$title",
						"page": "$raw"
					}
				}
			]
		}
	}
	`

	j, err := NewJob("unittest", rule)
	if err != nil {
		t.Fatalf("NewJob error: %v", err)
	}

	job = j
}
