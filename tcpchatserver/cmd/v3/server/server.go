// NATS based chat server. Can horizontally scale.
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
)

var (
	users = make(map[string]net.Conn)
	mu    sync.RWMutex
)

type Payload struct {
	IsAction bool
	User     net.Conn
	Message  string
}

type PayloadReceive struct {
	IsAction bool
	User     string
	Message  string
}

func main() {
	// Connect Options.
	opts := []nats.Option{nats.Name("NATS Sample Chat")}

	// Connect to NATS
	nc, err := nats.Connect("localhost", opts...)
	if err != nil {
		log.Fatal(err)
	}
	defer nc.Close()

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
			for _, user := range users {
				fmt.Fprintf(user, "Terminating server..\n")
				err := user.Close()
				check(err)
			}
			mu.RUnlock()
			os.Exit(0)
		}
	}()

	go broadcastmsg(nc)

	for {
		conn, err := a.Accept()
		check(err)
		go handleconn(nc, conn)
	}
}

func handleconn(nc *nats.Conn, conn net.Conn) {
	defer conn.Close()

	publishmsg(nc, Payload{
		IsAction: true,
		User:     conn,
		Message:  "joined",
	})

	bf := bufio.NewScanner(conn)
	for bf.Scan() {
		text := bf.Text()
		if len(text) == 0 {
			continue
		}

		publishmsg(nc, Payload{
			IsAction: false,
			User:     conn,
			Message:  text,
		})
	}

	publishmsg(nc, Payload{
		IsAction: true,
		User:     conn,
		Message:  "left",
	})
}

func broadcastmsg(nc *nats.Conn) {
	subj := "main-chat-topic"
	sub, _ := nc.Subscribe(subj, func(msg *nats.Msg) {

		payload := PayloadReceive{}
		err := json.Unmarshal(msg.Data, &payload)
		check(err)
		if payload.IsAction {
			switch string(payload.Message) {
			case "joined":
				mu.RLock()
				for _, i := range users {
					fmt.Fprintf(i, "%v joined\n", payload.User)
				}
				mu.RUnlock()
			case "left":
				mu.RLock()
				for _, i := range users {
					fmt.Fprintf(i, "%v left\n", payload.User)
				}
				mu.RUnlock()
			}
		} else {
			mu.RLock()
			for _, i := range users {
				fmt.Fprintf(i, "%v: %v\n", payload.User, payload.Message)
			}
			mu.RUnlock()
		}
	})
	//sub.SetPendingLimits(1000, 5*1024*1024)
	sub.SetPendingLimits(-1, -1)

	nc.Flush()
}

func publishmsg(nc *nats.Conn, payload Payload) {
	userId := fmt.Sprintf("%v", payload.User.RemoteAddr())

	if payload.IsAction {
		switch payload.Message {
		case "joined":
			mu.Lock()
			users[userId] = payload.User
			mu.Unlock()
		case "left":
			mu.Lock()
			delete(users, userId)
			mu.Unlock()
		}
	}

	payloadtosend := PayloadReceive{
		User:    userId,
		Message: payload.Message,
	}

	b, err := json.Marshal(payloadtosend)
	check(err)

	subj, msg := "main-chat-topic", b

	nc.Publish(subj, msg)
	nc.Flush()

	if err := nc.LastError(); err != nil {
		panic(err)
	}
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
