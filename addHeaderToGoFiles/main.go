package main

import (
	"bytes"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

func main() {

	header := `// HEADER
// HERE

`
	var bufferPool = sync.Pool{
		New: func() interface{} {
			return new(bytes.Buffer)
		},
	}

	if err := filepath.WalkDir(".", func(path string, d fs.DirEntry, err error) error {
		if strings.HasSuffix(path, ".go") && !d.IsDir() {
			f1buffer := bufferPool.Get().(*bytes.Buffer)
			defer bufferPool.Put(f1buffer)

			_, err := f1buffer.WriteString(header)
			if err != nil {
				return err
			}

			f2, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0666)
			if err != nil {
				return err
			}
			defer f2.Close()

			_, err = io.Copy(f1buffer, f2)
			if err != nil {
				return err
			}

			_, err = f2.Seek(0, io.SeekStart)
			if err != nil {
				return err
			}

			_, err = io.Copy(f2, f1buffer)
			if err != nil {
				return err
			}
		}

		return nil
	}); err != nil {
		panic(err)
	}
}
