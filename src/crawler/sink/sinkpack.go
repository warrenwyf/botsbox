package sink

import (
	"encoding/json"
)

type SinkPack struct {
	Name string
	Url  string
	Data map[string]interface{}
	File []byte
}

func (self *SinkPack) GetDataAsJson() (string, error) {
	bytes, err := json.Marshal(self.Data)
	return string(bytes), err
}
