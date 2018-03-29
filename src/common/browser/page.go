package browser

import (
	"errors"
	"time"

	"github.com/gotk3/gotk3/glib"

	"./webkit"
)

type Page struct {
	webView *webkit.WebView

	loadChan chan struct{}
	loadErr  error

	loaded bool
	closed bool
}

func (self *Page) Load(url string, timeout time.Duration) error {
	if self.closed {
		return errors.New("Page already closed")
	}

	self.loadChan = make(chan struct{})
	self.loadErr = nil
	self.loaded = false

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

	close(self.loadChan)

	return self.loadErr
}

func (self *Page) GetTitle() string {
	if self.closed || !self.loaded {
		return ""
	}

	c := make(chan string)

	glib.IdleAdd(func() bool {
		c <- self.webView.GetTitle()

		return false
	})

	title := <-c
	close(c)

	return title
}

func (self *Page) ExportMHtml(timeout time.Duration) []byte {
	if self.closed || !self.loaded {
		return nil
	}

	c := make(chan []byte)
	cClosed := false

	glib.IdleAdd(func() bool {
		self.webView.ExportMHtml(func(bytes []byte) {
			if !cClosed {
				c <- bytes
			}
		})

		return false
	})

	var mhtml []byte = nil

	t := time.NewTimer(timeout)

	select {
	case <-t.C:
	case mhtml = <-c:
		t.Stop()
	}

	cClosed = true
	close(c)

	return mhtml
}

func (self *Page) Close() {
	self.closed = true

	glib.IdleAdd(func() bool {
		self.webView.Destroy()

		return false
	})
}
