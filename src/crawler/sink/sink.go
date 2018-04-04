package sink

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"../../runtime"
	"../../store"
	"../../xlog"
)

type Sink struct {
	C chan *SinkPack
}

func NewSink() *Sink {
	return &Sink{
		C: make(chan *SinkPack, 1000),
	}
}

func (self *Sink) Open() {
	go self.loop()
}

func (self *Sink) loop() {
	s := store.GetStore()

	for {
		select {
		case sinkPack := <-self.C:

			if len(sinkPack.Data) > 0 { // Save to store
				datasetName := sinkPack.Name
				if !s.HasDataset(datasetName) {
					err := s.CreateDataset(datasetName,
						[]string{"id", "hash", "data", "createdAt"},
						[]string{"text", "text", "text", "timestamp DEFAULT CURRENT_TIMESTAMP"})
					if err != nil {
						xlog.Errln("Create dataset", datasetName, "error:", err)
					}
				}

				jsonData, errMarshal := json.Marshal(sinkPack.Data)
				if errMarshal != nil {
					xlog.Errln("SinkPack marshal error:", errMarshal)
				}

				_, err := s.InsertObject(datasetName,
					[]string{"id", "hash", "data"},
					[]interface{}{sinkPack.Id, sinkPack.Hash, jsonData})
				if err != nil {
					xlog.Errln("Insert into dataset", datasetName, "error:", err)
				}
			}

			if len(sinkPack.File) > 0 { // Save to file system
				dirName := sinkPack.Name
				dirPath := path.Join(runtime.GetAbsDataDir(), dirName)
				os.MkdirAll(dirPath, 0755)

				fileName := sinkPack.Id
				fileName = strings.Replace(fileName, string(os.PathSeparator), "_", -1)
				fileName = strings.Replace(fileName, string(os.PathListSeparator), "_", -1)

				ext := strings.ToLower(filepath.Ext(fileName))
				if ext != strings.ToLower(sinkPack.FileExt) {
					fileName += ext
				}

				filePath := path.Join(dirPath, fileName)
				if _, errStat := os.Stat(filePath); errStat == nil {
					xlog.Errln("File", filePath, "exists")
				} else {
					err := ioutil.WriteFile(filePath, sinkPack.File, 0755)
					if err != nil {
						xlog.Errln("Write file", filePath, "error:", err)
					}
				}
			}

		}
	}
}
