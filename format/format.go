package main

import (
	"bytes"
	"fmt"
	"go/format"
	"go/parser"
	"go/token"
	"log"
)

const expr = "(6+2*3)/4"

func main() {
	node, err := parser.ParseExpr(expr)
	if err != nil {
		log.Fatal(err)
	}

	fset := token.NewFileSet()

	var buf bytes.Buffer
	err = format.Node(&buf, fset, node)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(buf.String())
	source, err := format.Source([]byte(buf.String()))
	fmt.Println(string(source))
	fmt.Println(err)
}
