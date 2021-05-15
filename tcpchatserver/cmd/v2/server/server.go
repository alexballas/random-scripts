package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"time"

	"github.com/streadway/amqp"
)

var (
	users = make(map[net.Conn]bool)
)

type Say struct {
	user    net.Conn
	message string
}

type Do struct {
	user    net.Conn
	message string
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
			fmt.Println("Current Users:", len(users))
		}

	}()
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	go func() {
		for range c {
			for user := range users {
				fmt.Fprintf(user, "Terminating server..\n")
				err := user.Close()
				check(err)
			}
			os.Exit(0)

		}
	}()

	go broadcastmsg(messagesq, ch)
	go broadcastaction(actionq, ch)
	//go broadcastaction(queues)
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

	publishaction(action, channel, Do{
		user:    conn,
		message: "join",
	})

	bf := bufio.NewScanner(conn)
	for bf.Scan() {
		text := bf.Text()
		if len(text) == 0 {
			continue
		}
		publishmsg(msg, channel, Say{
			user:    conn,
			message: text,
		})
	}

	publishaction(action, channel, Do{
		user:    conn,
		message: "left",
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
		log.Printf("Received a message: %s", d.Body)
	}
}

func broadcastaction(action amqp.Queue, channel *amqp.Channel) {
	//	var mu sync.Mutex
	msgs, err := channel.Consume(
		action.Name, // queue
		"",          // consumer
		true,        // auto-ack
		false,       // exclusive
		false,       // no-local
		false,       // no-wait
		nil,         // args
	)
	check(err)

	for d := range msgs {
		log.Printf("Received action: %s", d.Body)
	}
	/*
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
		}*/
}

func publishmsg(msg amqp.Queue, channel *amqp.Channel, say Say) {

	err := channel.Publish(
		"",       // exchange
		msg.Name, // routing key
		false,    // mandatory
		false,    // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(say.message),
		})
	check(err)
}

func publishaction(action amqp.Queue, channel *amqp.Channel, do Do) {

	err := channel.Publish(
		"",          // exchange
		action.Name, // routing key
		false,       // mandatory
		false,       // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(do.message),
		})
	check(err)
}
