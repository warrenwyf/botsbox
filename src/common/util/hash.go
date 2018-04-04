package util

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
)

func Md5(v interface{}) string {
	h := md5.New()
	bytes, _ := json.Marshal(v)

	h.Write(bytes)
	return hex.EncodeToString(h.Sum(nil))
}

func Md5Bytes(bytes []byte) string {
	h := md5.New()
	h.Write(bytes)
	return hex.EncodeToString(h.Sum(nil))
}
