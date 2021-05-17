package main

import (
	"fmt"
	"net"
	"sync"
)

func main() {
	var wg sync.WaitGroup
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go fn0(&wg)
	}
	wg.Wait()

}

func fn0(wg *sync.WaitGroup) {
	defer wg.Done()
	conn, err := net.Dial("tcp", "localhost:12345")
	check(err)
	defer conn.Close()
	for i := 0; i < 100; i++ {
		fmt.Fprintf(conn, "Hi\n")
	}
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
