package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/chromedp"
)

func main() {
start:
	log.Println("go !")
	// create chrome instance
	ctx, cancel := chromedp.NewContext(
		context.Background(),
		chromedp.WithLogf(log.Printf),
	)
	defer cancel()

	// create a timeout
	ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	start := time.Now()
	// navigate to a page, wait for an element, click
	var res string
	err := chromedp.Run(ctx,
		emulation.SetUserAgentOverride("WebScraper 1.0"),
		chromedp.Navigate(`https://www.eopyykmes.gr/login.xhtml?viewId=/index.xhtml`),
		// wait for footer element is visible (ie, page is loaded)
		//chromedp.WaitVisible(`.form-signin`),

		chromedp.WaitVisible(`.container`),
		chromedp.ActionFunc(func(context.Context) error {
			log.Printf(">>>>>>>>>>>>>>>>>>>> BOX1 IS VISIBLE")
			return nil
		}))
	if err != nil {
		log.Println(err)
		goto start
	}
	fmt.Printf("h1 contains: '%s'\n", res)
	fmt.Printf("\nTook: %f secs\n", time.Since(start).Seconds())
}
