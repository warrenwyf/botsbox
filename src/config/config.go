package config

import (
	"encoding/json"
	"log"
	"os"
	"sync"
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
	file, errOpen := os.Open(filePath)
	if errOpen != nil {
		log.Println("Cannot read config file:", filePath, ",", errOpen.Error())
		return errOpen
	}

	decoder := json.NewDecoder(file)
	errDecode := decoder.Decode(&conf)
	if errDecode != nil {
		log.Println("Cannot parse config file:", filePath, ",", errDecode.Error())
		return errDecode
	}

	return nil
}

func GetConf() *Conf {
	once.Do(func() {
		confSingleton = &Conf{ // Default configuration
			HttpPort:  6075,
			StoreType: "sqlite",
			StoreConn: "./botsbox.db",
		}
	})

	return confSingleton
}
