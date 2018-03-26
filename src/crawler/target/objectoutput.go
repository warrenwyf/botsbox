package target

import (
	"github.com/tidwall/gjson"
)

type ObjectOutput struct {
	Name string
	Data map[string]string
}

func NewObjectOutput() *ObjectOutput {
	return &ObjectOutput{
		Data: map[string]string{},
	}
}

func NewObjectOutputWithJson(elem *gjson.Result) *ObjectOutput {
	output := NewObjectOutput()

	nameElem := elem.Get("$name")
	if nameElem.Exists() {
		output.Name = nameElem.String()
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
