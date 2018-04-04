package rule

import (
	"fmt"
	"strings"
)

func ApplyVarToString(s string, varName string, varValue string) string {
	placeholder := fmt.Sprintf(`$var[%s]`, varName)
	return strings.Replace(s, placeholder, varValue, -1)
}

func ApplyVarToMap(m map[string]string, varName string, varValue string) map[string]string {
	newM := map[string]string{}
	for k, v := range m {
		newM[k] = ApplyVarToString(v, varName, varValue)
	}

	return newM
}
