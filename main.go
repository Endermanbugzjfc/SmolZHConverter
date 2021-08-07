package main

import (
	"encoding/json"
	"github.com/andlabs/ui"
	"github.com/atotto/clipboard"
	"io/ioutil"
	"net/http"
	"net/url"
	"sync/atomic"
	"time"
)

const (
	Title = "【簡化雞】"
)

var (
	UserAuto atomic.Value
	Lock     atomic.Value
	Auto     *ui.Checkbox
	Entry    *ui.Entry
	Button   *ui.Button
)

func main() {
	UserAuto.Store(false)
	Lock.Store(false)
	go func() {
		for {
			if Auto == nil || Entry == nil || Lock.Load().(bool) {
				continue
			}
			ua := UserAuto.Load().(bool)
			if clipboard.Unsupported {
				if Auto.Checked() {
					ui.QueueMain(func() {
						Auto.SetChecked(false)
						Entry.Enable()
					})
				}
				if Auto.Enabled() {
					ui.QueueMain(func() {
						Auto.Disable()
					})
				}
			} else if !Auto.Enabled() {
				ui.QueueMain(func() {
					Auto.Enable()
				})
				if ua {
					ui.QueueMain(func() {
						Auto.SetChecked(true)
						Entry.Disable()
					})
				}
			}
		}
	}()
	err := ui.Main(func() {
		win := ui.NewWindow(Title+"翡翠出品。正宗廢品", 300, 30, true)
		win.OnClosing(func(*ui.Window) bool {
			Lock.Store(true)
			ui.Quit()
			return true
		})
		ui.OnShouldQuit(func() bool {
			win.Destroy()
			return true
		})

		box := ui.NewHorizontalBox()
		win.SetChild(box)

		Auto = ui.NewCheckbox("From Clipboard")
		box.Append(Auto, false)
		Auto.OnToggled(func(auto *ui.Checkbox) {
			UserAuto.Store(auto.Checked())
			if auto.Checked() {
				Entry.Disable()
			} else {
				Entry.Enable()
			}
		})

		Entry = ui.NewEntry()
		box.Append(Entry, false)

		Button = ui.NewButton("Magical Button")
		box.Append(Button, false)

		Button.OnClicked(func(button *ui.Button) {
			Button.Disable()
			var text string
			if UserAuto.Load().(bool) {
				all, err := clipboard.ReadAll()
				if err != nil {
					win.SetTitle(Title + err.Error())
				}
				text = all
			} else {
				text = Entry.Text()
			}

			dots := "."
			title := Title + text + " ➠ "
			win.SetTitle(title + dots)
			lock := true
			queue := func() {
				lock = false
				ui.QueueMain(func() {
					dots = dots + "."
					win.SetTitle(title + dots)
					if len(dots) >= 6 {
						dots = ""
					}
					lock = true
				})
			}

			result := make(chan string)
			rErr := make(chan string)
			ticker := time.NewTicker(time.Second / 2)

			go func() {
				for {
					select {
					case <-ticker.C:
						for !lock {
						}
						queue()
					case r := <-result:
						ticker.Stop()
						ui.QueueMain(func() {
							win.SetTitle(title + r)
							Button.Enable()
						})
						return
					case r := <-rErr:
						ticker.Stop()
						ui.QueueMain(func() {
							win.SetTitle(Title + r)
							Button.Enable()
						})
						return
					}
				}
			}()

			go func() { // Fanhuaji API
				client := http.Client{Timeout: time.Second * 30}
				r, err := client.Get("https://api.zhconvert.org/convert?text=" + url.QueryEscape(text) + "&converter=China")
				//goland:noinspection GoUnhandledErrorResult
				defer r.Body.Close()
				if err != nil {
					rErr <- err.Error()
					return
				}
				if r.StatusCode != http.StatusOK {
					rErr <- "Conversion server error"
					return
				}
				bodyBytes, err := ioutil.ReadAll(r.Body)
				if err != nil {
					rErr <- err.Error()
				}
				cr := &ConversionServerResponse{}
				err1 := json.Unmarshal(bodyBytes, cr)
				if err1 != nil {
					rErr <- err1.Error()
				}
				err2 := clipboard.WriteAll(cr.Data.Text)
				if err1 != nil {
					rErr <- err2.Error()
				}
				result <- cr.Data.Text
			}()
		})

		win.Show()
	})
	if err != nil {
		panic(err)
	}
}

type ConversionServerResponse struct {
	Data struct {
		Text string
	}
}
