package target

import (
	"github.com/tidwall/gjson"
)

type Output struct {
	Name string
	Data map[string]string
}

func NewOutputWithJson(elem gjson.Result) *Output {
	output := &Output{
		Data: map[string]string{},
	}

	nameElem := elem.Get("$name")
	if nameElem.Exists() {
		output.Name = nameElem.String()
	}

	dataElem := elem.Get("$data")
	if dataElem.Exists() {
		mapElem := dataElem.Map()
		if mapElem != nil {
			for k, v := range mapElem {
				output.Data[k] = v.String()
			}
		}
	}

	return output
}
