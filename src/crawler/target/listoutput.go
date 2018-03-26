package target

import (
	"github.com/tidwall/gjson"
)

type ListOutput struct {
	Name     string
	Selector string
	Data     map[string]string
}

func NewListOutput() *ListOutput {
	return &ListOutput{
		Data: map[string]string{},
	}
}

func NewListOutputWithJson(elem *gjson.Result) *ListOutput {
	output := NewListOutput()

	nameElem := elem.Get("$name")
	if nameElem.Exists() {
		output.Name = nameElem.String()
	}

	eachElem := elem.Get("$each")
	if eachElem.Exists() {
		output.Selector = eachElem.String()
	}

	dataElem := elem.Get("$data")
	if dataElem.Exists() {
		dataElem.ForEach(func(kElem, vElem gjson.Result) bool {
			output.Data[kElem.String()] = vElem.String()
			return true
		})
	}

	return output
}
