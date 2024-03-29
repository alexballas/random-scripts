package main

import (
	"bufio"
	"bytes"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gen2brain/dlgs"
)

func main() {
	file, _, err := dlgs.File("Select file", "", false)
	if err != nil {
		dlgs.Error("Error", err.Error())
		panic(err)
	}

	numberstring, _, err := dlgs.Entry("Entry", "Enter desired split value... xrodroulh:", "")
	if err != nil {
		dlgs.Error("Error", err.Error())
		panic(err)
	}

	number, err := strconv.Atoi(numberstring)
	if err != nil {
		dlgs.Error("Error", "Xodroulh... mallon ekanes malakia. Prepei na baleis noumeraki")
		panic(err)
	}

	if number == 1 {
		dlgs.Error("Error", "Xodroulh... bale kati megalytero apo 1")
		panic(err)
	}

	file2, err := os.Open(file)
	if err != nil {
		dlgs.Error("Error", err.Error())
		panic(err)
	}

	newfile, err := os.Create("nice.txt")
	if err != nil {
		dlgs.Error("Error", err.Error())
		panic(err)
	}

	bufioo := bufio.NewScanner(file2)
	realcounter := 0
	buf := new(bytes.Buffer)

	for bufioo.Scan() {
		if bytes.ContainsRune(bufioo.Bytes(), '(') {
			continue
		} else {
			realcounter++
		}
		if realcounter%number == 1 {
			buf.Reset()
		}
		if realcounter%number == 0 {
			buf.Write(bufioo.Bytes())
			buf.WriteString("\r\n")
			newfile.Write(buf.Bytes())
			buf.Reset()
			continue
		}
		buf.Write(bufioo.Bytes())
	}
	newfile.Write(buf.Bytes())

	stringpath, err := filepath.Abs("nice.txt")
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
