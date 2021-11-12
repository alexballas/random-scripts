package main

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stdout, "no file defined\n")
		os.Exit(1)
	}

	f, err := os.Open(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stdout, "%s\n", err.Error())
		os.Exit(1)
	}

	bytes, _ := io.ReadAll(f)

	build := new(strings.Builder)

	for _, q := range bytes {
		el := strconv.FormatInt(int64(q), 16)
		if len(el) == 1 {
			build.WriteString("0x0")
			build.WriteString(strings.ToUpper(el) + `, `)
			continue
		}

		build.WriteString(`0x` + strings.ToUpper(el) + `, `)
	}

	fmt.Println(`[]byte{` + strings.TrimRight(build.String(), `, `) + `}`)
}
