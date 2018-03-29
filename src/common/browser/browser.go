package browser

import (
	"errors"
	"runtime"
	"sync"

	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"

	"./webkit"
)

var (
	browserSingleton *Browser
	once             sync.Once

	sig struct{}
)

type Browser struct {
	C chan *Page
}

func GetBrowser() *Browser {
	once.Do(func() {
		browserSingleton = &Browser{
			C: make(chan *Page),
		}

		go browserSingleton.loop()
	})

	return browserSingleton
}

func (self *Browser) loop() {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	gtk.Init(nil)

	gtk.Main()
}

func (self *Browser) Destroy() {
	gtk.MainQuit()

	close(self.C)
}

func (self *Browser) CreatePage() *Page {
	glib.IdleAdd(func() bool {
		defer func() {
			if err := recover(); err != nil {
				self.C <- nil
			}
		}()

		webView := webkit.NewWebView()

		if webView != nil {
			settings := webView.GetSettings()
			settings.SetAutoLoadImages(false) // Do not load images by default

			page := &Page{
				webView: webView,
				closed:  false,
			}

			widget := webView.GetGtk()

			loadChangedHandler, _ := widget.Connect("load-changed",
				func(_ *glib.Object, event int) {
					defer func() {
						if err := recover(); err != nil {
							page.loadErr = errors.New("Page load chagned panic")
						}
					}()

					if event == webkit.LoadEvent_Finished {
						if page.closed {
							page.loadErr = errors.New("Page already closed")
						} else {
							page.loaded = true
						}

						page.loadChan <- sig
					}
				})

			widget.Connect("load-failed",
				func() {
					defer func() {
						if err := recover(); err != nil {
							page.loadErr = errors.New("Page load failed panic")
						}
					}()

					widget.HandlerDisconnect(loadChangedHandler)

					if page.closed {
						page.loadErr = errors.New("Page already closed")
					} else {
						page.loadErr = errors.New("Page load failed")
					}

					page.loadChan <- sig
				})

			self.C <- page
		} else {
			self.C <- nil
		}

		return false
	})

	return <-self.C
}
