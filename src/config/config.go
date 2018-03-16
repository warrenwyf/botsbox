package config

import (
	"encoding/json"
	"log"
	"os"
	"sync"
)

const (
	VersionMajor int = 0
	VersionMinor int = 0
	VersionPatch int = 1
)

var (
	confSingleton *Conf
	once          sync.Once
)

type Conf struct {
	HttpPort  int    `json:"http_port"`
	StoreType string `json:"store_type"`
	StoreConn string `json:"store_conn"`
	StoreName string `json:"store_name"`
}

func (conf *Conf) SyncFromFile(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		log.Println("Cannot read config file:", filePath, ", Error:", err.Error())
		return err
	}

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&conf)
	if err != nil {
		log.Println("Cannot parse config file:", filePath, ", Error:", err.Error())
		return err
	}

	return nil
}

func GetConf() *Conf {
	once.Do(func() {
		confSingleton = &Conf{ // Default configuration
			HttpPort: 6075,
		}
	})

	return confSingleton
}
