package sink

type SinkPack struct {
	Name    string
	Hash    string
	Url     string
	Data    map[string]interface{}
	File    []byte
	FileExt string
}
