package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gen2brain/dlgs"
)

func main() {
	var deksia []string
	weekly_values := make(map[int][]string)

	file, _, err := dlgs.File("Select file", "", false)
	if err != nil {
		dlgs.Error("Error", err.Error())
		panic(err)
	}

	data, err := os.Open(file)
	if err != nil {
		dlgs.Error("Error", err.Error())
		panic(err)
	}

	newfile, err := os.Create("nice2.txt")
	if err != nil {
		dlgs.Error("Error", err.Error())
		panic(err)
	}

	buf := bufio.NewScanner(data)
	counter := 0
	for buf.Scan() {
		counter++
		if counter == 1 {
			deksia = strings.Fields(buf.Text())
			continue
		}
		weekly_values[counter-2] = strings.Fields(buf.Text())
		if counter == 8 {
			break
		}
	}
	var buffer bytes.Buffer
	for position, actual := range deksia {
		for _, qq := range weekly_values[position%7] {
			toint_actual, err := strconv.ParseFloat(actual, 64)
			if err != nil {
				dlgs.Error("Error", err.Error())
				panic(err)
			}
			toint, err := strconv.ParseFloat(qq, 64)
			if err != nil {
				dlgs.Error("Error", err.Error())
				panic(err)
			}

			result := toint_actual * toint
			fmt.Println(result, toint_actual, toint)
			buffer.WriteString(fmt.Sprintf("%f", result))
			buffer.WriteString("\r\n")
		}
	}
	newfile.Write(buffer.Bytes())

	stringpath, err := filepath.Abs("nice2.txt")
	if err != nil {
		dlgs.Error("Error", err.Error())
		panic(err)
	}
	_, err = dlgs.Info("Info", stringpath)
	if err != nil {
		dlgs.Error("Error", err.Error())
		panic(err)
	}
}
