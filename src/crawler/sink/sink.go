package sink

import (
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

func (self *Sink) Open(store store.Store) {
	go self.loop(store)
}

func (self *Sink) loop(store store.Store) {
	for {
		select {
		case sinkPack := <-self.C:

			datasetName := sinkPack.Name
			if !store.HasDataset(datasetName) {
				err := store.CreateDataset(datasetName,
					[]string{"url", "data", "createdAt"},
					[]string{"text", "text", "timestamp DEFAULT CURRENT_TIMESTAMP"})
				if err != nil {
					xlog.Errln("Create dataset", datasetName, "error:", err)
				}
			}

			jsonData, errMarshal := sinkPack.GetDataAsJson()
			if errMarshal != nil {
				xlog.Errln("SinkPack marshal error:", errMarshal)
			}

			_, err := store.InsertObject(datasetName,
				[]string{"url", "data"},
				[]interface{}{sinkPack.Url, jsonData})
			if err != nil {
				xlog.Errln("Insert into dataset", datasetName, "error:", err)
			}

		}
	}
}
