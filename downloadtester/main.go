package main

import (
	"fmt"
	"io"
	"net/http"
	"sync/atomic"
	"time"
)

func main() {
	tik := time.NewTicker(1 * time.Second)
	var start uint64

	req, err := http.Get("https://ftp.bit.nl/pub/OpenBSD/7.8/amd64/install78.img")
	if err != nil {
		panic(err)
	}
	defer req.Body.Close()

	count := &Counter{req.Body, 0}
	go func() {
		for range tik.C {
			fmt.Printf("\r%0.3f MiB/s", float64((atomic.LoadUint64(&count.bytes)-start))/float64((1024*1024)))
			start = atomic.LoadUint64(&count.bytes)
		}
	}()

	io.Copy(io.Discard, count)

	fmt.Println("\nTotal bytes transfered: ", atomic.LoadUint64(&count.bytes))
}

type Counter struct {
	io.Reader
	bytes uint64
}

func (c *Counter) Read(b []byte) (int, error) {
	n, err := c.Reader.Read(b)
	atomic.AddUint64(&c.bytes, uint64(n))
	return n, err
}
