package browser

import (
	"errors"
	"runtime"
	"time"

	"github.com/gotk3/gotk3/glib"

	"./webkit"
)

type Page struct {
	webView *webkit.WebView

	loadChan       chan struct{}
	loadChanClosed bool

	loadErr error
	loaded  bool

	closed bool
}

func (self *Page) Load(url string, timeout time.Duration) error {
	if self.closed {
		return errors.New("Page already closed")
	}

	self.loadChan = make(chan struct{})
	self.loadChanClosed = false

	glib.IdleAdd(func() bool {
		if self.closed {
			self.loadErr = errors.New("Page already closed")
		} else {
			self.webView.LoadURI(url)
		}

		return false
	})

	t := time.NewTimer(timeout)

	select {
	case <-t.C:
		self.loadErr = errors.New("Page load timeout")
	case <-self.loadChan:
		t.Stop()
	}

	self.loadChanClosed = true
	close(self.loadChan)

	return self.loadErr
}

func (self *Page) GetTitle() string {
	if self.closed || !self.loaded {
		return ""
	}

	c := make(chan string)
	cClosed := false

	glib.IdleAdd(func() bool {
		if !cClosed {
			c <- self.webView.GetTitle()
		}

		return false
	})

	title := <-c

	cClosed = true
	close(c)

	return title
}

func (self *Page) ExportMHtml(timeout time.Duration) []byte {
	if self.closed || !self.loaded {
		return nil
	}

	c := make(chan []byte)

	var callbackHolder interface{}

	glib.IdleAdd(func() bool {
		callbackHolder = self.webView.ExportMHtml(c)

		return false
	})

	var mhtml []byte = nil

	t := time.NewTimer(timeout)

	select {
	case <-t.C:
	case mhtml = <-c:
		t.Stop()
	}

	runtime.KeepAlive(callbackHolder) // Avoid GC() callback

	close(c)

	return mhtml
}

func (self *Page) Close() {
	self.closed = true

	glib.IdleAdd(func() bool {
		self.webView.Destroy()
		self.webView = nil

		return false
	})
}
