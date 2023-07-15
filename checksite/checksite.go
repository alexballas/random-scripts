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

	/* 	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", false),
	) */

	allocContext, cancel := chromedp.NewRemoteAllocator(context.Background(), "ws://127.0.0.1:9222/devtools/browser/b806f75d-353d-49ca-95b9-55096618da8b")
	defer cancel()
	/*
		allocCtx, cancel := chromedp.NewExecAllocator(allocatorContext, opts...)
		defer cancel() */

	ctx, cancel2 := chromedp.NewContext(allocContext)
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
			time.Sleep(10 * time.Second)
			return nil
		}))

	if err != nil {
		log.Println(err)
		goto start
	}
	fmt.Printf("\nTook: %f secs\n", time.Since(start).Seconds())
}
