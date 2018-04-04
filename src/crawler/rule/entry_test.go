package rule

import (
	"fmt"
	"testing"

	"github.com/tidwall/gjson"
)

func Test_Parameterized(t *testing.T) {
	json := `
	{
		"$name": "recommend_api",
		"$url": "https://site",
		"$method": "post",
		"$form": {
			"params": "{\"limit\":20,\"offset\":$var[offset]}"
		},
		"$contentType": "json",
		"$var": {
			"offset": "$[0, 20, 40, 60, 80, 100]",
			"offset": "$rangeInt[0, 100, 50]"
		}
	}
	`

	elem := gjson.Parse(json)
	entry := NewEntryWithJson(&elem)

	if entry.Name != "recommend_api" {
		t.Fatal("Name error:", entry.Name)
	}

	if entry.Url != "https://site" {
		t.Fatal("Url error:", entry.Url)
	}

	if entry.Method != "POST" {
		t.Fatal("Method error:", entry.Method)
	}

	params := entry.Form["params"]
	if params != "{\"limit\":20,\"offset\":$var[offset]}" {
		t.Fatal("Form error:", entry.Form)
	}

	if entry.ContentType != "json" {
		t.Fatal("ContentType error:", entry.ContentType)
	}

	offsetVar, _ := entry.Var["offset"]
	if fmt.Sprintf("%v", offsetVar) != fmt.Sprintf("%v", []string{"0", "50", "100"}) {
		t.Fatal(entry.Var["offset"], "Does not mean:", offsetVar)
	}
}
