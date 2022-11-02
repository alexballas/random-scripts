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

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", false),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancel2 := chromedp.NewContext(
		allocCtx,
		chromedp.WithLogf(log.Printf),
	)
	defer cancel2()

	// create a timeout
	ctx, cancel3 := context.WithTimeout(ctx, 10*time.Second)
	defer cancel3()

	start := time.Now()
	// navigate to a page, wait for an element, click
	err := chromedp.Run(ctx,
		emulation.SetUserAgentOverride("WebScraper 1.0"),
		chromedp.Navigate(`https://www.eopyykmes.gr/login.xhtml?viewId=/index.xhtml`),
		chromedp.WaitVisible(`.container`),
		chromedp.ActionFunc(func(context.Context) error {
			log.Printf(">>>>>>>>>>>>>>>>>>>> BOX1 IS VISIBLE")
			return nil
		}))

	if err != nil {
		log.Println(err)
		goto start
	}
	fmt.Printf("\nTook: %f secs\n", time.Since(start).Seconds())
}
