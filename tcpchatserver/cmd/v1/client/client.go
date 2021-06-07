package main

import (
	"fmt"
	"net"
	"sync"
)

func main() {
	var wg sync.WaitGroup
	for i := 0; i < 4000; i++ {
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

	fmt.Fprintf(conn, "Hi\n")
	fmt.Fprintf(conn, "Hello\n")

}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
