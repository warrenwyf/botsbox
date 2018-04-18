package crawler

import (
	"bufio"
	"bytes"
	"math/rand"
	"os"
	"strings"
	"sync"

	"../runtime"
)

var (
	userAgents []string
	once       sync.Once
)

func getUserAgents() []string {
	once.Do(func() {
		filePath := runtime.GetAbsUaFile()

		file, errOpen := os.Open(filePath)
		if errOpen != nil {
			return
		}
		defer file.Close()

		reader := bufio.NewReader(file)
		buf := bytes.Buffer{}

		for {
			line, isPrefix, err := reader.ReadLine()
			if len(line) > 0 {
				buf.Write(line)
				if !isPrefix {
					str := strings.TrimSpace(buf.String())
					if len(str) > 0 {
						userAgents = append(userAgents, str)
					}

					buf.Reset()
				}
			}

			if err != nil {
				break
			}
		}
	})

	return userAgents
}

func RandomUserAgent() string {
	userAgents := getUserAgents()

	size := len(userAgents)
	if size > 0 {
		idx := rand.Intn(size)
		return userAgents[idx]
	}

	return ""
}
