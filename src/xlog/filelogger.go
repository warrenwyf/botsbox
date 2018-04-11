package xlog

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

const (
	defaultBufferSize    = 4096
	defaultChanSize      = 1000
	defaultFlushInterval = time.Duration(5) * time.Second
)

var signal = struct{}{}

type FileLogger struct {
	fileName   string
	bufferSize int

	logChan   chan string
	flushChan chan struct{}
	closeChan chan struct{}

	buf          *bytes.Buffer
	file         *os.File
	fileOpenedAt time.Time
}

func NewFileLogger(fileName string) *FileLogger {
	return NewFileLoggerWithOptions(fileName, defaultBufferSize, defaultChanSize, defaultFlushInterval)
}

func NewFileLoggerWithOptions(fileName string, bufferSize int, chanSize int, flushInterval time.Duration) *FileLogger {
	dirPath := path.Dir(fileName)
	os.MkdirAll(dirPath, 0755)

	file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0755)
	if err != nil || file == nil {
		return nil
	}

	if bufferSize < 0 {
		bufferSize = defaultBufferSize
	}

	if chanSize < 0 {
		chanSize = defaultChanSize
	}

	l := &FileLogger{
		fileName:   fileName,
		bufferSize: bufferSize,

		logChan:   make(chan string, chanSize),
		flushChan: make(chan struct{}),
		closeChan: make(chan struct{}),

		buf:          &bytes.Buffer{},
		file:         file,
		fileOpenedAt: time.Now(),
	}

	fileInfo, errStat := os.Stat(fileName)
	if errStat == nil {
		stat := fileInfo.Sys().(*syscall.Stat_t)
		createdAt := time.Unix(int64(stat.Birthtimespec.Sec), int64(stat.Birthtimespec.Nsec))
		l.checkAndRotateFile(createdAt)
	}

	if flushInterval <= 0 {
		flushInterval = defaultFlushInterval
	}
	go l.loop(flushInterval)

	return l
}

func (self *FileLogger) Printf(format string, v ...interface{}) {
	str := fmt.Sprintf(format, v...)

	self.print(str)
}

func (self *FileLogger) Println(v ...interface{}) {
	str := fmt.Sprintln(v...)

	self.print(str)
}

func (self *FileLogger) Flush() {
	defer func() {
		recover()
	}()

	self.flushChan <- signal
}

func (self *FileLogger) Close() {
	defer func() {
		recover()
	}()

	self.closeChan <- signal
}

func (self *FileLogger) print(str string) {
	defer func() {
		recover()
	}()

	self.logChan <- str
}

func (self *FileLogger) loop(flushInterval time.Duration) {
	var ticker = time.NewTicker(flushInterval)
	defer ticker.Stop()

	defer func() {
		close(self.logChan)
		close(self.flushChan)
		close(self.closeChan)
	}()

	for {
		select {
		case <-ticker.C:
			self.writeFile()
			self.file.Sync()

		case str := <-self.logChan:
			self.writeBuffer(str)
			if self.buf.Len() > self.bufferSize {
				self.writeFile()
			}

		case <-self.flushChan:
			self.writeFile()
			self.file.Sync()

		case <-self.closeChan:
			self.file.Close()
			goto end
		}
	}

end:
}

func (self *FileLogger) writeBuffer(str string) error {
	now := time.Now()
	log := fmt.Sprintf(`%s %s`, now.Format(time.RFC3339), str)

	_, err := self.buf.WriteString(log)
	return err
}

func (self *FileLogger) writeFile() error {
	b := self.buf.Bytes()
	if len(b) == 0 {
		return nil
	} else {
		self.buf.Reset()
	}

	_, errWrite := self.file.Write(b)
	if errWrite != nil {
		return errWrite
	}

	errRotate := self.checkAndRotateFile(self.fileOpenedAt)
	if errRotate != nil {
		return errRotate
	}

	return nil
}

func (self *FileLogger) checkAndRotateFile(t time.Time) error {
	now := time.Now()
	sameDay := (now.Year() == t.Year() && now.Month() == t.Month() && now.Day() == t.Day())
	if !sameDay {
		errRotate := self.rotateFile()
		if errRotate != nil {
			return errRotate
		}
	}

	return nil
}

func (self *FileLogger) rotateFile() error {
	self.file.Sync()
	self.file.Close()

	ext := filepath.Ext(self.fileName)
	name := strings.TrimSuffix(self.fileName, ext)
	stamp := time.Now().Format("20060102")
	newFileName := fmt.Sprintf(`%s_%s%s`, name, stamp, ext)

	if _, err := os.Stat(newFileName); err == nil {
		return errors.New(fmt.Sprintf(`File "%s" already exists`, newFileName))
	}

	errRename := os.Rename(self.fileName, newFileName)
	if errRename != nil {
		return errRename
	}

	file, errOpen := os.OpenFile(self.fileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if errOpen != nil {
		return errOpen
	}

	self.file = file
	self.fileOpenedAt = time.Now()

	return nil
}
