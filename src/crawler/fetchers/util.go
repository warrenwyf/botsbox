package fetchers

import (
	"fmt"
	"strings"
)

func joinQueryString(params map[string]string) string {
	strs := []string{}

	for k, v := range params {
		strs = append(strs, fmt.Sprintf("%s=%s", k, v))
	}

	return strings.Join(strs, "&")
}
