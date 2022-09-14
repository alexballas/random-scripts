package main

import (
	"fmt"
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
	triggerButtonApp := app.New()
	triggerButtonApp.Settings().SetTheme(theme.DarkTheme())
	mainWindow := triggerButtonApp.NewWindow("Trigger Button")
	status := widget.NewLabel("")
	button = widget.NewButton("CLICK ME", func() {
		go trigger(status)
	})
	text := container.New(layout.NewHBoxLayout(), layout.NewSpacer(), status, layout.NewSpacer())
	content := container.New(layout.NewVBoxLayout(), layout.NewSpacer(), button, text, layout.NewSpacer())
	mainWindow.SetContent(content)

	mainWindow.ShowAndRun()

}

func trigger(status *widget.Label) {
	go clearStatus(status)

	button.Disable()
	status.Text = ""
	status.Refresh()

	defer button.Enable()
	defer status.Refresh()

	client := http.Client{
		Timeout: time.Duration(10 * time.Second),
	}

	req, err := http.NewRequest("GET", "http://192.168.88.240/trigger", nil)

	if err != nil {
		status.Text = err.Error()
		return
	}
	req.Host = "wifibutton1"
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
		clearTimer.Stop()
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
