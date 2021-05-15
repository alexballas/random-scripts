// RabbitMQ based chat server
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/streadway/amqp"
)

var (
	users = make(map[string]net.Conn)
	mu    sync.RWMutex
)

type Payload struct {
	User    net.Conn
	Message string
}

type PayloadReceive struct {
	User    string
	Message string
}

func main() {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	check(err)
	defer conn.Close()

	ch, err := conn.Channel()
	check(err)
	defer ch.Close()

	messagesq, err := ch.QueueDeclare(
		"messages", // name
		false,      // durable
		false,      // delete when unused
		false,      // exclusive
		false,      // no-wait
		nil,        // arguments
	)
	check(err)

	actionq, err := ch.QueueDeclare(
		"actions", // name
		false,     // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	check(err)

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

	go broadcastmsg(messagesq, ch)
	go broadcastaction(actionq, ch)
	for {
		conn, err := a.Accept()
		check(err)
		go handleconn(conn, actionq, messagesq, ch)

	}
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func handleconn(conn net.Conn, action amqp.Queue, msg amqp.Queue, channel *amqp.Channel) {
	defer conn.Close()

	publishaction(action, channel, Payload{
		User:    conn,
		Message: "joined",
	})

	bf := bufio.NewScanner(conn)
	for bf.Scan() {
		text := bf.Text()
		if len(text) == 0 {
			continue
		}
		publishmsg(msg, channel, Payload{
			User:    conn,
			Message: text,
		})
	}

	publishaction(action, channel, Payload{
		User:    conn,
		Message: "left",
	})

}

func broadcastmsg(msg amqp.Queue, channel *amqp.Channel) {
	//	var mu sync.Mutex
	msgs, err := channel.Consume(
		msg.Name, // queue
		"",       // consumer
		true,     // auto-ack
		false,    // exclusive
		false,    // no-local
		false,    // no-wait
		nil,      // args
	)
	check(err)

	for d := range msgs {
		say := PayloadReceive{}
		err := json.Unmarshal(d.Body, &say)
		check(err)
		mu.RLock()
		for _, i := range users {
			fmt.Fprintf(i, "%v: %v\n", say.User, say.Message)
		}
		mu.RUnlock()
	}
}

func broadcastaction(action amqp.Queue, channel *amqp.Channel) {
	actions, err := channel.Consume(
		action.Name, // queue
		"",          // consumer
		true,        // auto-ack
		false,       // exclusive
		false,       // no-local
		false,       // no-wait
		nil,         // args
	)
	check(err)

	for d := range actions {
		action := PayloadReceive{}
		err := json.Unmarshal(d.Body, &action)
		check(err)

		switch string(action.Message) {
		case "joined":
			mu.RLock()
			for _, i := range users {
				fmt.Fprintf(i, "%v joined\n", action.User)
			}
			mu.RUnlock()
		case "left":
			mu.RLock()
			for _, i := range users {
				fmt.Fprintf(i, "%v left\n", action.User)
			}
			mu.RUnlock()
		}
	}
}

func publishmsg(msg amqp.Queue, channel *amqp.Channel, payload Payload) {
	saytosend := PayloadReceive{
		User:    fmt.Sprintf("%v", payload.User.RemoteAddr()),
		Message: payload.Message,
	}

	b, err := json.Marshal(saytosend)
	check(err)

	err = channel.Publish(
		"",       // exchange
		msg.Name, // routing key
		false,    // mandatory
		false,    // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        b,
		})
	check(err)
}

func publishaction(action amqp.Queue, channel *amqp.Channel, payload Payload) {
	userId := fmt.Sprintf("%v", payload.User.RemoteAddr())

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

	actiontosend := PayloadReceive{
		User:    fmt.Sprintf("%v", payload.User.RemoteAddr()),
		Message: payload.Message,
	}

	b, err := json.Marshal(actiontosend)
	check(err)

	err = channel.Publish(
		"",          // exchange
		action.Name, // routing key
		false,       // mandatory
		false,       // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        b,
		})
	check(err)
}
