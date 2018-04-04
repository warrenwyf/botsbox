package rule

import (
	"strconv"
	"strings"

	"github.com/tidwall/gjson"
)

type Entry struct {
	Name       string
	Url        string
	Method     string
	Header     map[string]string
	Query      map[string]string
	Form       map[string]string
	ResultType string
	Var        map[string][]string
}

func NewEntryWithJson(elem *gjson.Result) *Entry {
	entry := &Entry{
		Method:     "GET",
		Header:     map[string]string{},
		Query:      map[string]string{},
		Form:       map[string]string{},
		ResultType: "html",
		Var:        map[string][]string{},
	}

	nameElem := elem.Get("$name")
	if nameElem.Exists() {
		entry.Name = nameElem.String()
	}

	urlElem := elem.Get("$url")
	if urlElem.Exists() {
		entry.Url = urlElem.String()
	}

	methodElem := elem.Get("$method")
	if methodElem.Exists() {
		entry.Method = strings.ToUpper(methodElem.String())
	}

	headerElem := elem.Get("$header")
	if headerElem.Exists() {
		headerElem.ForEach(func(kElem, vElem gjson.Result) bool {
			entry.Header[kElem.String()] = vElem.String()
			return true
		})
	}

	queryElem := elem.Get("$query")
	if queryElem.Exists() {
		queryElem.ForEach(func(kElem, vElem gjson.Result) bool {
			entry.Query[kElem.String()] = vElem.String()
			return true
		})
	}

	formElem := elem.Get("$form")
	if formElem.Exists() {
		formElem.ForEach(func(kElem, vElem gjson.Result) bool {
			entry.Form[kElem.String()] = vElem.String()
			return true
		})
	}

	resultTypeElem := elem.Get("$resultType")
	if resultTypeElem.Exists() {
		entry.ResultType = strings.ToLower(resultTypeElem.String())
	}

	varElem := elem.Get("$var")
	if varElem.Exists() {
		varElem.ForEach(func(kElem, vElem gjson.Result) bool {
			entry.Var[kElem.String()] = parseVar(vElem.String())
			return true
		})
	}

	return entry
}

/**
 * $[0, 20, 40, 60, 80, 100]
 * $rangeInt[0, 100, 20]
 */
func parseVar(def string) []string {
	lower := strings.ToLower(def)

	if strings.HasPrefix(lower, "$[") && strings.HasSuffix(lower, "]") {
		// List
		str := strings.TrimSuffix(strings.TrimPrefix(lower, "$["), "]")
		strs := strings.Split(str, ",")
		for i, str := range strs {
			strs[i] = strings.TrimSpace(str)
		}
		return strs

	} else if strings.HasPrefix(lower, "$rangeint[") && strings.HasSuffix(lower, "]") {
		// Int range
		str := strings.TrimSuffix(strings.TrimPrefix(lower, "$rangeint["), "]")
		strs := strings.Split(str, ",")
		for i, str := range strs {
			strs[i] = strings.TrimSpace(str)
		}

		count := len(strs)
		if count < 2 {
			return strs
		} else {
			from, errFrom := strconv.Atoi(strs[0])
			to, errTo := strconv.Atoi(strs[1])
			if errFrom != nil || errTo != nil {
				return strs
			}

			step := 1
			if count > 2 {
				v, vErr := strconv.Atoi(strs[2])
				if vErr == nil {
					step = v
				}
			}

			intStrs := []string{}
			for i := from; i <= to; i += step {
				intStrs = append(intStrs, strconv.Itoa(i))
			}

			return intStrs
		}

	}

	return []string{def}
}
