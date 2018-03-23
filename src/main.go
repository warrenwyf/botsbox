package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"./config"
	"./runtime"
	"./server"
)

func main() {
	version := fmt.Sprintf("%d.%d.%d", runtime.VersionMajor, runtime.VersionMinor, runtime.VersionPatch)
	if len(os.Args) > 1 {
		if os.Args[1] == "-v" || os.Args[1] == "--version" {
			fmt.Println("botsbox version:", version)
			return
		} else if os.Args[1] == "-h" {
			fmt.Println("# Usage: botsbox [options]")
			fmt.Println("  options:")
			fmt.Println("    -h                 Show help")
			fmt.Println("    -v, --version      Print version")
			fmt.Println("    -c, --config       Configuration file path")
			fmt.Println("    -d, --data         Data directory path")
			fmt.Println("    -l, --log          Log directory path")
			return
		}
	}

	var (
		defaultConfigFile = "./config.json"
		usageConfigFile   = "Configuration file path"

		defaultDataDir = "./data"
		usageDataDir   = "Data directory path"

		defaultLogDir = "./log"
		usageLogDir   = "Log directory path"
	)

	flag.StringVar(&runtime.ConfigFile, "c", defaultConfigFile, usageConfigFile)
	flag.StringVar(&runtime.ConfigFile, "config", defaultConfigFile, usageConfigFile)
	flag.StringVar(&runtime.DataDir, "d", defaultDataDir, usageDataDir)
	flag.StringVar(&runtime.DataDir, "data", defaultDataDir, usageDataDir)
	flag.StringVar(&runtime.LogDir, "l", defaultLogDir, usageLogDir)
	flag.StringVar(&runtime.LogDir, "log", defaultLogDir, usageLogDir)
	flag.Parse()

	// Get current working directory path
	runtime.WorkingDir, _ = filepath.Abs(filepath.Dir(os.Args[0]))

	// Read configuration file
	conf := config.GetConf()
	errConf := conf.SyncFromFile(runtime.GetAbsConfigFile())
	if errConf != nil {
		fmt.Println("⚡  Server will be started with default options")
	}

	err := server.Start()
	if err != nil {
		fmt.Printf("🔥  botsbox %s failed to start: %v \n", version, err)
		fmt.Println("\t Configuration file path: ", runtime.ConfigFile)
		fmt.Println("\t Data directory path: ", runtime.DataDir)
		fmt.Println("\t Log directory path: ", runtime.LogDir)
	}
}
