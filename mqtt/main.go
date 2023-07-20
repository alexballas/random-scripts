package main

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"strconv"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func handleMessage(client mqtt.Client, msg mqtt.Message) {
	num, err := strconv.Atoi(string(msg.Payload()))
	if err != nil {
		return
	}

	switch {
	case num >= 70:
		log.Println("High", num)
	case num < 70 && num >= 30:
		log.Println("Mid", num)
	case num < 30:
		log.Println("Low", num)
	}
}

func main() {
	uri, err := url.Parse("http://192.168.88.107:1883/shellies/button1/sensor/battery")
	if err != nil {
		log.Fatal(err)
	}
	topic := uri.Path[1:len(uri.Path)]

	mqtt.ERROR = log.New(os.Stdout, "ERROR ", log.Ldate|log.Ltime)
	mqtt.CRITICAL = log.New(os.Stdout, "CRITICAL ", log.Ldate|log.Ltime)

	opts := mqtt.NewClientOptions()
	opts.SetMaxReconnectInterval(30 * time.Second)
	opts.SetOnConnectHandler(func(client mqtt.Client) {
		if token := client.Subscribe(topic, 0, handleMessage); token.Wait() && token.Error() == nil {
			fmt.Println("Subscribed to", topic)
		}
	})

	opts.AddBroker(fmt.Sprintf("tcp://%s", uri.Host))

	mqtt.NewClient(opts).Connect()

	select {}
}
