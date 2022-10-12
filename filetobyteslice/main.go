package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

func main() {
	var genericFile *os.File

	fi, err := os.Stdin.Stat()
	var notFile bool

	if err == nil && fi.Mode()&os.ModeNamedPipe != 0 {
		notFile = true
		genericFile = os.Stdin
	}

	if !notFile {
		if len(os.Args) < 2 {
			fmt.Fprintf(os.Stdout, "no file defined\n")
			os.Exit(1)
		}

		f, err := os.Open(os.Args[1])
		if err != nil {
			fmt.Fprintf(os.Stdout, "%s\n", err.Error())
			os.Exit(1)
		}
		defer f.Close()

		genericFile = f
	}

	build := new(strings.Builder)

	if notFile && len(os.Args) == 2 && os.Args[1] == "assembly" {
		buf := bufio.NewScanner(os.Stdin)
		for buf.Scan() {
			t := strings.TrimSpace(buf.Text())
			tSlice := strings.Fields(t)
			if len(tSlice) > 2 && tSlice[0][:2] == "0x" {
				build.WriteString(tSlice[1])
			}
		}

		finalBuild := new(strings.Builder)

		var temp string
		for index, num := range build.String() {
			if index%2 != 0 {
				a := temp + string(num)
				finalBuild.WriteString("0x" + a + ", ")
				continue
			}

			temp = string(num)

		}
		finalString := `[]byte{` + strings.TrimRight(finalBuild.String(), ", ") + `}`
		fmt.Println(finalString)
		os.Exit(0)
	}
	bytes, _ := io.ReadAll(genericFile)

	for _, q := range bytes {
		el := strconv.FormatInt(int64(q), 16)
		if len(el) == 1 {
			build.WriteString(`0x0` + strings.ToUpper(el) + `, `)
			continue
		}

		build.WriteString(`0x` + strings.ToUpper(el) + `, `)
	}

	fmt.Println(`[]byte{` + strings.TrimRight(build.String(), `, `) + `}`)
}
