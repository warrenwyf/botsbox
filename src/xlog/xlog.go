package xlog

import (
	"fmt"
	"os"
	"path"
	"sync"

	"../runtime"
)

var (
	outLogger Logger
	errLogger Logger

	outOnce sync.Once
	errOnce sync.Once
)

type Logger interface {
	Printf(format string, v ...interface{})
	Println(v ...interface{})
	Flush()
	Close()
}

func Outf(format string, v ...interface{}) {
	getOutLogger().Printf(format, v...)
}

func Outln(v ...interface{}) {
	getOutLogger().Println(v...)
}

func Errf(format string, v ...interface{}) {
	getErrLogger().Printf(format, v...)
}

func Errln(v ...interface{}) {
	getErrLogger().Println(v...)
}

func FlushAll() {
	getOutLogger().Flush()
	getErrLogger().Flush()
}

func CloseAll() {
	getOutLogger().Close()
	getErrLogger().Close()
}

func getOutLogger() Logger {
	outOnce.Do(func() {
		fileName := path.Join(runtime.GetAbsLogDir(), "botsbox-out.log")
		outLogger = NewFileLogger(fileName)

		if outLogger == nil {
			fmt.Println("Output logger can not be null")
			os.Exit(1)
		}
	})

	return outLogger
}

func getErrLogger() Logger {
	errOnce.Do(func() {
		fileName := path.Join(runtime.GetAbsLogDir(), "botsbox-err.log")
		errLogger = NewFileLogger(fileName)

		if errLogger == nil {
			fmt.Println("Error logger can not be null")
			os.Exit(1)
		}
	})

	return errLogger
}
