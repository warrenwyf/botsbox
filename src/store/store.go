package store

import (
	"os"
	"path"

	"../config"
	"../runtime"
)

const (
	JobDataset = "$job"
)

type Store interface {
	Init() error

	HasDataset(datasetName string) bool
	CreateDataset(datasetName string, fieldNames []string, fieldTypes []string) error

	InsertObject(dataset string, fields []string, values []interface{}) (oid string, err error)
	DeleteObjects(dataset string, oids []string) (count int64, err error)

	QueryAllJobs() (jobs []map[string]interface{}, err error)

	Destroy() error
}

func NewStore() Store {
	conf := config.GetConf()

	if conf.StoreType == "mongo" {
	}

	// Default use sqlite for dev & test
	filePath := conf.StoreConn
	if !path.IsAbs(filePath) {
		dataDir := runtime.GetAbsDataDir()
		os.MkdirAll(dataDir, 0755)
		filePath = path.Join(dataDir, filePath)
	}
	return &SqliteStore{
		FilePath: filePath,
	}
}
