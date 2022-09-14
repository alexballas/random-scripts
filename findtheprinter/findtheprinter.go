package main

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/go-retryablehttp"
)

func main() {
	found := false
	var foundCurIP string
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	for i := 1; i < 255; i++ {
		iChar := strconv.Itoa(i)

		go func(ctx context.Context) {
			retryClient := retryablehttp.NewClient()
			retryClient.RetryMax = 5
			retryClient.HTTPClient.Timeout = 3 * time.Second
			retryClient.Logger = nil
			client := retryClient.StandardClient()

			curIP := "http://192.168.2." + iChar

			req, err := http.NewRequestWithContext(ctx, http.MethodGet, curIP, nil)
			if err != nil {
				return
			}

			resp, err := client.Do(req)
			if err != nil {
				return
			}
			defer resp.Body.Close()

			buf := bufio.NewScanner(resp.Body)
			for buf.Scan() {
				if bytes.Contains(buf.Bytes(), []byte(`SEIKO EPSON`)) {
					foundCurIP = curIP
					found = true
					cancel()
					return
				}
			}
		}(ctx)
	}

	<-ctx.Done()

	if !found {
		log.Printf("Issue %s\n", errors.New("no IP found").Error())
		os.Exit(1)
	}

	foundCurIP = strings.Replace(foundCurIP, "http://", "", -1)
	foundCurIPb := []byte(foundCurIP)
	var currentIP []byte

	f, err := os.ReadFile("/etc/cups/printers.conf")
	if err != nil {
		log.Printf("Issue %s\n", err)
		os.Exit(1)
	}

	buf := bufio.NewScanner(bytes.NewReader(f))

	for buf.Scan() {
		if bytes.Contains(buf.Bytes(), []byte(`DeviceURI lpd://192.168.2`)) {
			currentIP = getCurrentIP(buf.Bytes())
			break
		}
	}

	if bytes.Equal(currentIP, foundCurIPb) {
		log.Println("SAME IP, ALL GOOD!")
		os.Exit(0)
	}

	log.Println("Current IP: ", string(currentIP))
	log.Println("New IP: ", string(foundCurIPb))

	newFileBytes := bytes.ReplaceAll(f, currentIP, foundCurIPb)

	err = os.WriteFile("/etc/cups/printers.conf", newFileBytes, 0600)
	if err != nil {
		log.Printf("Issue %s\n", err)
		os.Exit(1)
	}

	_, err = exec.Command("/bin/systemctl", "restart", "cups").Output()
	if err != nil {
		log.Printf("Issue %s\n", err)
		os.Exit(1)
	}
}

func getCurrentIP(b []byte) []byte {
	element := []byte("DeviceURI lpd://")
	element1 := []byte(":515/PASSTHRU")

	first := bytes.Replace(b, element, []byte{}, -1)
	second := bytes.Replace(first, element1, []byte{}, -1)
	return second
}
