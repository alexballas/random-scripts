package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func connect(clientId string, uri *url.URL) mqtt.Client {
	opts := createClientOptions(clientId, uri)
	client := mqtt.NewClient(opts)
	token := client.Connect()
	for !token.WaitTimeout(3 * time.Second) {
	}
	if err := token.Error(); err != nil {
		log.Fatal(err)
	}
	return client
}

func createClientOptions(clientId string, uri *url.URL) *mqtt.ClientOptions {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s", uri.Host))
	return opts
}

func listen(uri *url.URL, topic string) {
	client := connect("sub", uri)
	client.Subscribe(topic, 0, func(client mqtt.Client, msg mqtt.Message) {
		temps, _ := getTemp()
		log.Printf("* [%s] %s %s", msg.Topic(), string(msg.Payload()), temps)
	})
}

func main() {
	uri, err := url.Parse("http://192.168.88.2:1883/ac")
	if err != nil {
		log.Fatal(err)
	}
	topic := uri.Path[1:len(uri.Path)]

	f, err := os.OpenFile("ac-temps.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()
	log.SetOutput(f)

	go listen(uri, topic)
	select {}
}

func getTemp() (string, error) {
	client := http.Client{}
	req, err := http.NewRequest("GET", "http://192.168.88.3/", nil)
	if err != nil {
		return "", err
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	body = bytes.ReplaceAll(body, []byte("<br>"), []byte(" "))

	return string(body), nil
}
