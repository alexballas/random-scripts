// RabbitMQ based chat server. Can horizontally scale.
// This is just a proof of concept. You can easily deploy a
// RabbitMQ instance by running the following:
// docker run -it --rm --name rabbitmq -p 5672:5672 -p 15672:15672 rabbitmq:3-management
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
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	check(err)
	defer conn.Close()

	ch, err := conn.Channel()
	check(err)
	defer ch.Close()

	ch2, err := conn.Channel()
	check(err)
	defer ch2.Close()

	// Configured based on
	// https://www.rabbitmq.com/tutorials/tutorial-three-go.html
	// Will be used only for receiving
	err = ch.ExchangeDeclare(
		"msgexchange", // name
		"fanout",      // type
		true,          // durable
		false,         // auto-deleted
		false,         // internal
		false,         // no-wait
		nil,           // arguments
	)
	check(err)

	// Will be used only for publishing
	err = ch2.ExchangeDeclare(
		"msgexchange", // name
		"fanout",      // type
		true,          // durable
		false,         // auto-deleted
		false,         // internal
		false,         // no-wait
		nil,           // arguments
	)
	check(err)

	messagesq, err := ch.QueueDeclare(
		"",    // name
		false, // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	check(err)

	err = ch.QueueBind(
		messagesq.Name, // queue name
		"",             // routing key
		"msgexchange",  // exchange
		false,
		nil,
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

	for {
		conn, err := a.Accept()
		check(err)
		go handleconn(conn, ch2)
	}
}

func handleconn(conn net.Conn, channel *amqp.Channel) {
	defer conn.Close()

	publishmsg(channel, Payload{
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

		publishmsg(channel, Payload{
			IsAction: false,
			User:     conn,
			Message:  text,
		})
	}

	publishmsg(channel, Payload{
		IsAction: true,
		User:     conn,
		Message:  "left",
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
		payload := PayloadReceive{}
		err := json.Unmarshal(d.Body, &payload)
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
	}
}

func publishmsg(channel *amqp.Channel, payload Payload) {
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

	err = channel.Publish(
		"msgexchange", // exchange
		"",            // routing key
		false,         // mandatory
		false,         // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        b,
		})
	check(err)
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
