package sink

type SinkPack struct {
	Name string

	Id      string
	Hash    string
	Data    map[string]interface{}
	File    []byte
	FileExt string
}
