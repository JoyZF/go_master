package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
)

func main() {
	ReadFile("/Users/joy/workspace/gopath/src/go_master/read_big_file/1.txt", func(bytes []byte) {
		fmt.Println(string(bytes))
	})
}

// ReadFile read file by bufio
func ReadFile(filePath string, handle func([]byte)) error {
	f, err := os.Open(filePath)
	defer f.Close()
	if err != nil {
		panic(err)
	}

	buf := bufio.NewReader(f)
	for {
		line, _, err := buf.ReadLine()
		handle(line)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
	}
}

func ReadBigFile(fileName string, handle func([]byte)) error {
	f, err := os.Open(fileName)
	defer f.Close()
	if err != nil {
		panic(err)
	}
	return nil
}
