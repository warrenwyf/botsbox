package xlog

import (
	"path"
	"testing"
	"time"
)

var logFileName = path.Join("/tmp", "botsbox-test.log")

func Test_FileLogger(t *testing.T) {
	logger := NewFileLogger(logFileName)
	if logger == nil {
		t.Fatalf(`NewFileLogger(%s) failed`, logFileName)
	}

	// Default flush interval is 5 sec, only 5 logs will be saved
	for i := 0; i < 7; i++ {
		logger.Println("This is a log", i)
		time.Sleep(time.Second)
	}

	logger.Close()

	logger.Println("This can not be logged")
	logger.Flush()
}

func Test_FileLogger_Flush(t *testing.T) {
	logger := NewFileLogger(logFileName)
	if logger == nil {
		t.Fatalf(`NewFileLogger(%s) failed`, logFileName)
	}

	logger.Println("Flush x log")
	time.Sleep(time.Second)
	logger.Println("Flush y log")

	logger.Flush()
	logger.Close()
}

func Test_FileLogger_rotateFile(t *testing.T) {
	logger := NewFileLogger(logFileName)
	if logger == nil {
		t.Fatalf(`NewFileLogger(%s) failed`, logFileName)
	}

	err := logger.rotateFile()
	if err != nil {
		t.Logf("FileLogger.rotateFile() error: %v", err)
	}

	logger.Close()
}
