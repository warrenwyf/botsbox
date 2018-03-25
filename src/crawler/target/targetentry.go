package target

import (
	"github.com/tidwall/gjson"
)

type Entry struct {
	Name string
	Url  string
}

func NewEntryWithJson(elem gjson.Result) *Entry {
	entry := &Entry{}

	nameElem := elem.Get("$name")
	if nameElem.Exists() {
		entry.Name = nameElem.String()
	}

	urlElem := elem.Get("$url")
	if urlElem.Exists() {
		entry.Url = urlElem.String()
	}

	return entry
}
