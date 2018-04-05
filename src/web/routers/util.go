package routers

import (
	"encoding/json"
	"fmt"

	"github.com/labstack/echo"
)

func joinPath(prefix string, path string) string {
	return fmt.Sprintf("%s%s", prefix, path)
}

func writeJsonResponse(resp *echo.Response, v interface{}) error {
	b, errMarshal := json.Marshal(v)
	if errMarshal != nil {
		return errMarshal
	}

	_, err := resp.Write(b)
	return err
}
