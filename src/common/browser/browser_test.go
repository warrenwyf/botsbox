package browser

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/signal"
	"runtime"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	"../mhtml"
)

var (
	sigChan chan os.Signal
	server  *httptest.Server
)

func setup() {
	sigChan = make(chan os.Signal)
	signal.Notify(sigChan, syscall.SIGABRT, syscall.SIGSEGV, syscall.SIGBUS, syscall.SIGILL)

	go func() {
		for {
			select {
			case sig := <-sigChan:
				fmt.Println("System signal catched: ", sig)
			}
		}
	}()

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(res http.ResponseWriter, req *http.Request) {
		title := req.URL.Query().Get("title")

		html := fmt.Sprintf("<html><head><title>%s</title></head></html>", title)
		res.Write([]byte(html))
	})

	server = httptest.NewServer(mux)
}

func teardown() {
	server.Close()

	close(sigChan)
}

func TestMain(m *testing.M) {
	setup()

	m.Run()

	teardown()
}

func Test_Browser_Concurrency(t *testing.T) {
	count := 100

	var noTitleCount uint64 = 0
	var noHtmlCount uint64 = 0

	c := make(chan *Page)

	for i := 0; i < count; i++ {

		go func(idx int) {
			fmt.Printf("Page %d open at %v \n", idx, time.Now())

			p := GetBrowser().CreatePage()

			loadErr := p.Load(fmt.Sprintf("%s?title=%d", server.URL, idx), 5*time.Second)
			if loadErr != nil {
				fmt.Printf("Page %d Load() error at %v: %v\n", idx, time.Now(), loadErr)
			}

			title := p.GetTitle()
			if len(title) == 0 {
				atomic.AddUint64(&noTitleCount, 1)
			}
			fmt.Printf("Page %d GetTitle() returns \"%s\" at %v\n", idx, title, time.Now())

			html := mhtml.GetHtml(p.ExportMHtml(5 * time.Second))
			if len(html) == 0 {
				atomic.AddUint64(&noHtmlCount, 1)
			}
			fmt.Printf("Page %d export HTML \"%s\" at %v\n", idx, html, time.Now())

			p.Close()

			c <- p
		}(i)

	}

	finished := 0

	for {
		<-c

		runtime.GC() // Make sure GC() does not release unsafe pointers

		finished++
		if finished == count {
			break
		}
	}

	t.Logf(`%d/%d pages did not get title correctly`, noTitleCount, count)
	t.Logf(`%d/%d pages did not export html correctly`, noHtmlCount, count)
}

func Test_Browser_Persistent(t *testing.T) {
	count := 10

	var noTitleCount uint64 = 0
	var noHtmlCount uint64 = 0

	for idx := 0; idx < count; idx++ {
		fmt.Printf("Page %d open at %v \n", idx, time.Now())

		p := GetBrowser().CreatePage()

		loadErr := p.Load(fmt.Sprintf("%s?title=%d", server.URL, idx), 5*time.Second)
		if loadErr != nil {
			fmt.Printf("Page %d Load() error at %v: %v\n", idx, time.Now(), loadErr)
		}

		title := p.GetTitle()
		if len(title) == 0 {
			atomic.AddUint64(&noTitleCount, 1)
		}
		fmt.Printf("Page %d GetTitle() returns \"%s\" at %v\n", idx, title, time.Now())

		html := mhtml.GetHtml(p.ExportMHtml(5 * time.Second))
		if len(html) == 0 {
			atomic.AddUint64(&noHtmlCount, 1)
		}
		fmt.Printf("Page %d export HTML \"%s\" at %v\n", idx, html, time.Now())

		p.Close()
	}

	t.Logf(`%d/%d pages did not get title correctly`, noTitleCount, count)
	t.Logf(`%d/%d pages did not export html correctly`, noHtmlCount, count)
}

func Test_Browser_Destroy(t *testing.T) {
	GetBrowser().Destroy()
}
