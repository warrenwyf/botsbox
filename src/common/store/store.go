package store

import (
	"../../config"
)

const (
	JobDataset = "job"
)

type Store interface {
	Init() error

	CreateDataset(datasetName string, fieldNames []string, fieldTypes []string) (err error)

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
	return &SqliteStore{
		FilePath: conf.StoreConn,
	}
}
