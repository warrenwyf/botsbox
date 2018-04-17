package fetchers

import (
	"fmt"
	"net/http"
	"strings"
)

func joinQueryString(params map[string]string) string {
	strs := []string{}

	for k, v := range params {
		strs = append(strs, fmt.Sprintf("%s=%s", k, v))
	}

	return strings.Join(strs, "&")
}

/*
 * Get the first value associated with the given name, name is case insensitive
 */
func headerValueIgnoreCase(header http.Header, name string) string {
	for k, v := range header {
		if strings.ToLower(k) == strings.ToLower(name) && len(v) > 0 {
			return v[0]
		}
	}

	return ""
}
