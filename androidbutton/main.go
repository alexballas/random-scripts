package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

var (
	button         *widget.Button
	clearTimer     *time.Timer
	killgoroutines = make(chan struct{})
)

func main() {

	r := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{
				Timeout: time.Millisecond * time.Duration(10000),
			}
			return d.DialContext(ctx, network, "192.168.88.1:53")
		},
	}

	triggerButtonApp := app.New()
	triggerButtonApp.Settings().SetTheme(theme.DarkTheme())
	mainWindow := triggerButtonApp.NewWindow("Trigger Button")
	status := widget.NewLabel("")
	button = widget.NewButton("CLICK ME", func() {
		go trigger(r, status)
	})
	text := container.New(layout.NewHBoxLayout(), layout.NewSpacer(), status, layout.NewSpacer())
	content := container.New(layout.NewVBoxLayout(), layout.NewSpacer(), button, text, layout.NewSpacer())
	mainWindow.SetContent(content)

	mainWindow.ShowAndRun()

}

func trigger(r *net.Resolver, status *widget.Label) {
	go clearStatus(status)

	button.Disable()
	status.Text = ""
	status.Refresh()

	defer button.Enable()
	defer status.Refresh()

	timeOut, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	targetIPs, err := r.LookupHost(timeOut, "wifibutton1")
	if err != nil {
		status.Text = "DNS Failure"
		return
	}

	client := http.Client{
		Timeout: time.Duration(10 * time.Second),
	}

	req, err := http.NewRequest("GET", fmt.Sprintf("http://%s/trigger", targetIPs[0]), nil)
	req.Host = "wifibutton1"
	if err != nil {
		status.Text = err.Error()
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		status.Text = err.Error()
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		status.Text = resp.Status
		return
	}

	status.Text = fmt.Sprintf("OK %s", resp.Header.Get("action"))
}

func clearStatus(status *widget.Label) {
	if clearTimer != nil {
		killgoroutines <- struct{}{}
		if !clearTimer.Stop() {
			<-clearTimer.C
		}
	}

	clearTimer = time.NewTimer(15 * time.Second)
	select {
	case <-clearTimer.C:
		clearTimer = nil
		status.Text = ""
		status.Refresh()
	case <-killgoroutines:
		return
	}

}
