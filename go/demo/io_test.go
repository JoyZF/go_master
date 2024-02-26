package demo

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
)

func ExampleDemo() {
	var r io.Reader
	r = os.Stdin
	r = bufio.NewReader(r)
	r = new(bytes.Buffer)
	fmt.Println(r)
	// Output:
	//
}
