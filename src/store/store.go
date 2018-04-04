package store

import (
	"os"
	"path"
	"sync"

	"../config"
	"../runtime"
)

const (
	JobDataset    = "$job"
	TargetDataset = "$target"
)

var (
	storeSingleton Store
	once           sync.Once
)

type Store interface {
	Init() error

	HasDataset(datasetName string) bool
	CreateDataset(datasetName string, fieldNames []string, fieldTypes []string) error

	InsertObject(dataset string, fields []string, values []interface{}) (oid string, err error)
	DeleteObjects(dataset string, oids []string) (count int64, err error)

	QueryAllJobs() (jobs []map[string]interface{}, err error)
	GetLatestTarget(hash string) (target map[string]interface{}, err error)

	Destroy() error
}

func GetStore() Store {
	once.Do(func() {
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
		storeSingleton = &SqliteStore{
			FilePath: filePath,
		}
	})

	return storeSingleton
}
