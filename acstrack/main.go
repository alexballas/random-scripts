package main

import (
	"bufio"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
)

//go:embed beep.mp3
var efs embed.FS

type MainType struct {
	Items []Items `json:"items"`
	Count int     `json:"count"`
}

type StatusHistory struct {
	TrackingNumber      string `json:"trackingNumber"`
	ControlPointDate    string `json:"controlPointDate"`
	Description         string `json:"description"`
	ControlPoint        string `json:"controlPoint"`
	Info                string `json:"info"`
	ControlPointCode    string `json:"controlPointCode"`
	ControlPointFeature string `json:"controlPointFeature"`
}

type ProgressStatus struct {
	TrackingNumber string `json:"trackingNumber"`
	Description    string `json:"description"`
	Date           string `json:"date,omitempty"`
	Step           int    `json:"step"`
	IsCurrent      bool   `json:"isCurrent"`
}

type Items struct {
	PickupDate                        string           `json:"pickupDate"`
	ExpectedDeliveryDate              string           `json:"expectedDeliveryDate"`
	Sender                            string           `json:"sender"`
	RecipientName                     string           `json:"recipientName"`
	DestinationStationID              string           `json:"destinationStationID"`
	Notes                             string           `json:"notes"`
	SpecialFeatures                   string           `json:"specialFeatures"`
	RelatedTrackingNumber             string           `json:"relatedTrackingNumber"`
	TrackingNumber                    string           `json:"trackingNumber"`
	DestinationDescription            string           `json:"destinationDescription"`
	Recipient                         string           `json:"recipient"`
	NonDeliveryReason                 string           `json:"nonDeliveryReason"`
	DestinationStationTelephoneNumber string           `json:"destinationStationTelephoneNumber"`
	PickupStation                     string           `json:"pickupStation"`
	PickupDescription                 string           `json:"pickupDescription"`
	RecipientAddress                  string           `json:"recipientAddress"`
	PickupStationTelephoneNumber      string           `json:"pickupStationTelephoneNumber"`
	ProgressStatus                    []ProgressStatus `json:"progressStatus"`
	StatusHistory                     []StatusHistory  `json:"statusHistory"`
	DestinationRegion                 int              `json:"destinationRegion"`
	PickupRegion                      int              `json:"pickupRegion"`
	ShipmentStatusID                  int              `json:"shipmentStatusId"`
	DestinationPrefecture             int              `json:"destinationPrefecture"`
	PickupPrefecture                  int              `json:"pickupPrefecture"`
	IsDelivered                       bool             `json:"isDelivered"`
	IsReturned                        bool             `json:"isReturned"`
}

func main() {
	var key string
	var lines int

	stringb := new(strings.Builder)

	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Please pass the ACS tracking number\n")
		os.Exit(1)
	}

	trackingNumber, err := strconv.Atoi(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid ACS tracking number\n")
		os.Exit(1)
	}

	for {
		res, err := http.Get("https://www.acscourier.net/el/web/greece/track-and-trace")
		if err != nil {
			time.Sleep(time.Second)
			continue
		}

		buf := bufio.NewScanner(res.Body)
		for buf.Scan() {
			if strings.Contains(buf.Text(), "publicToken") {
				text := strings.TrimSpace(buf.Text())
				text = strings.TrimLeft(text, `publicToken="`)
				text = strings.TrimRight(text, `"`)
				key = text
				break
			}
		}

		io.Copy(io.Discard, res.Body)
		res.Body.Close()

		client := &http.Client{}
		req, err := http.NewRequest(http.MethodGet, "https://api.acscourier.net/api/parcels/search/"+strconv.Itoa(trackingNumber), nil)
		if err != nil {
			time.Sleep(time.Second)
			continue
		}
		req.Header.Set("authority", "api.acscourier.net")
		req.Header.Set("accept", "application/json")
		req.Header.Set("origin", "https://www.acscourier.net")
		req.Header.Set("referer", "https://www.acscourier.net/")
		req.Header.Set("x-encrypted-key", key)

		resp, err := client.Do(req)
		if err != nil {
			time.Sleep(time.Second)
			continue
		}

		var dataJSON MainType
		if err := json.NewDecoder(resp.Body).Decode(&dataJSON); err != nil {
			time.Sleep(time.Second)
			continue
		}
		resp.Body.Close()

		for _, item := range dataJSON.Items {
			for _, hist := range item.StatusHistory {
				stringb.WriteString(fmt.Sprintf("%s %s %s\n", hist.ControlPoint, hist.ControlPointDate, hist.Description))
			}
		}

		newLines := len(strings.Split(stringb.String(), "\n"))
		if newLines > lines {
			lines = newLines
			go func() {
				if err := playMP3Sound(); err != nil {
					fmt.Println(err)
				}
			}()

			log.Println()
			fmt.Println(stringb.String())
		}

		stringb.Reset()
		time.Sleep(15 * time.Minute)
	}
}

func playMP3Sound() error {
	f, err := efs.Open("beep.mp3")
	if err != nil {
		return err
	}
	defer f.Close()

	streamer, format, err := mp3.Decode(f)
	if err != nil {
		return err
	}
	defer streamer.Close()

	if err := speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10)); err != nil {
		return err
	}
	done := make(chan bool)

	speaker.Play(beep.Seq(streamer, beep.Callback(func() {
		done <- true
	})))

	<-done
	return nil
}
