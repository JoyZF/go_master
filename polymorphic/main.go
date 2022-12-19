package main

import "fmt"

type People interface {
	Name() string
}



type Child struct {

}

func NewChild() *Child {
	return &Child{}
}

func (c *Child) Name() string {
	return "child"
}

type Audit struct {

}

func NewAudit() *Audit {
	return &Audit{}
}

func (c *Audit) Name() string {
	return "audit"
}

func main()  {
	var a = 2
	var b People
	switch  {
	case a == 1:
		b = NewChild()
	case a == 2:
		b = NewAudit()
	}

	fmt.Println(b.Name())
}
