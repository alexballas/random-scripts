package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

type responseJSON []struct {
	Type  string `json:"vartype"`
	ID    string `json:"varid"`
	Value string `json:"varvalue"`
}

func main() {
	var toJSON responseJSON

	client := &http.Client{}

	req, err := http.NewRequest("GET", "http://192.168.1.1/data/Status.json", nil)
	check(err)

	req.Header = http.Header{
		"Accept-Language": {
			"el-GR,el;q=0.9,en;q=0.8",
		},
	}

	resp, err := client.Do(req)
	check(err)
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&toJSON)
	check(err)

	for _, q := range toJSON {
		if q.ID == "dsl_downstream" {
			fmt.Println("DSL Down:  ", q.Value)
		}
		if q.ID == "dsl_upstream" {
			fmt.Println("DSL Up:    ", q.Value)
		}
		if q.ID == "dsl_snr" {
			fmt.Println("SNR:       ", q.Value)
		}
		if q.ID == "vdsl_atnd" {
			fmt.Println("Att Down:  ", q.Value)
		}
		if q.ID == "vdsl_atnu" {
			fmt.Println("Att Up:    ", q.Value)
		}
		if q.ID == "dsl_crc_errors" {
			fmt.Println("CRC Errors:", q.Value)
		}
		if q.ID == "dsl_fec_errors" {
			fmt.Println("FEC Errors:", q.Value)
		}
	}
}

func check(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
