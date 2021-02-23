package main

import (
	"bufio"
	"bytes"
	"errors"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

func main() {
	found := false
	// we need at least capacity 1 since by
	// the time we find the printer we dont
	// want to block. The loop is probably still
	// going. These two factors seem to deadlock
	// our app.
	foundChannel := make(chan struct{}, 1)
	var foundCurIP string
	timer := time.After(20 * time.Second)

	for i := 1; i < 255; i++ {
		iChar := strconv.Itoa(i)

		go func() {
			try := 1
			client := &http.Client{
				Timeout: 10 * time.Second,
			}
			curIP := "http://192.168.2." + iChar
		AGAIN:
			resp, err := client.Get(curIP)

			if err != nil {
				if try < 20 {
					try++
					time.Sleep(1 * time.Second)
					goto AGAIN
				}
				return
			}

			defer resp.Body.Close()

			buf := bufio.NewScanner(resp.Body)

			for buf.Scan() {
				if bytes.Contains(buf.Bytes(), []byte(`SEIKO EPSON`)) {
					foundCurIP = curIP
					found = true
					foundChannel <- struct{}{}
					return
				}
			}
		}()
		if found {
			break
		}
	}

	select {
	case <-timer:
		break
	case <-foundChannel:
		break
	}

	if !found {
		log.Fatalf("Issue %s\n", errors.New("No IP found"))
		os.Exit(1)
	}
	foundCurIP = strings.Replace(foundCurIP, "http://", "", -1)
	foundCurIPb := []byte(foundCurIP)
	var currentIP []byte
	f, err := os.ReadFile("/etc/cups/printers.conf")
	if err != nil {
		log.Fatalf("Issue %s\n", err)
		os.Exit(1)
	}
	buf := bufio.NewScanner(bytes.NewReader(f))

	for buf.Scan() {
		if bytes.Contains(buf.Bytes(), []byte(`DeviceURI lpd://192.168.2`)) {
			currentIP = getCurrentIP(buf.Bytes())
		}
	}

	if bytes.Equal(currentIP, foundCurIPb) {
		log.Println("SAME IP, ALL GOOD!")
		os.Exit(0)
	}
	log.Println("Current IP: ", string(currentIP))
	log.Println("New IP: ", string(foundCurIPb))

	newFileBytes := bytes.ReplaceAll(f, currentIP, foundCurIPb)

	err = os.WriteFile("/etc/cups/printers.conf", newFileBytes, 600)
	if err != nil {
		log.Fatalf("Issue %s\n", err)
		os.Exit(1)
	}

	_, err = exec.Command("/bin/systemctl", "restart", "cups").Output()
	if err != nil {
		log.Fatalf("Issue %s\n", err)
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
