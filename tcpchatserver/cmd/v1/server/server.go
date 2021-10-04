package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"os/signal"
	"sync"
	"time"
)

var (
	msg    = make(chan Say)
	joined = make(chan net.Conn)
	left   = make(chan net.Conn)
	users  = make(map[net.Conn]bool)
	mu     sync.RWMutex
)

type Say struct {
	user    net.Conn
	message string
}

func main() {
	a, err := net.Listen("tcp", ":12345")
	check(err)

	ticker := time.NewTicker(time.Second * 5)

	go func() {
		for range ticker.C {
			mu.RLock()
			fmt.Println("Current Users:", len(users))
			mu.RUnlock()
		}

	}()
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	go func() {
		for range c {
			mu.RLock()
			for user := range users {
				fmt.Fprintf(user, "Terminating server..\n")
				err := user.Close()
				check(err)
			}
			mu.RUnlock()
			os.Exit(0)

		}
	}()

	go broadcast()
	for {
		conn, err := a.Accept()
		check(err)
		go handleconn(conn)

	}
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func handleconn(conn net.Conn) {
	defer conn.Close()

	joined <- conn

	bf := bufio.NewScanner(conn)
	for bf.Scan() {
		text := bf.Text()
		if len(text) == 0 {
			continue
		}
		msg <- Say{
			user:    conn,
			message: text,
		}
	}

	left <- conn

}

func broadcast() {
	for {
		select {
		case cur := <-msg:
			for i := range users {
				if i != cur.user {
					fmt.Fprintf(i, "%v: %v\n", cur.user, cur.message)
				}
			}
		case cjoin := <-joined:
			for i := range users {
				fmt.Fprintf(i, "%v joined\n", cjoin)
			}
			mu.Lock()
			users[cjoin] = true
			mu.Unlock()
		case cleft := <-left:
			for i := range users {
				fmt.Fprintf(i, "%v left\n", cleft)
			}
			mu.Lock()
			delete(users, cleft)
			mu.Unlock()
		}
	}
}
