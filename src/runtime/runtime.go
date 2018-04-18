package runtime

import (
	"path"
)

var (
	ConfigFile string
	UaFile     string

	DataDir    string
	LogDir     string
	WorkingDir string
)

func GetAbsConfigFile() string {
	if path.IsAbs(ConfigFile) {
		return ConfigFile
	}

	return path.Join(WorkingDir, ConfigFile)
}

func GetAbsUaFile() string {
	if path.IsAbs(UaFile) {
		return UaFile
	}

	return path.Join(WorkingDir, UaFile)
}

func GetAbsDataDir() string {
	if path.IsAbs(DataDir) {
		return DataDir
	}

	return path.Join(WorkingDir, DataDir)
}

func GetAbsLogDir() string {
	if path.IsAbs(LogDir) {
		return LogDir
	}

	return path.Join(WorkingDir, LogDir)
}
