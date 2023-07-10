package main

import "fmt"

type Good struct {
	Name  string
	Price int
}

func (g *Good) Sell() {
	fmt.Println("sell it")
}

type Book struct {
	Good
	Total int
}

//func (b *Book) Sell()  {
//	fmt.Println("sell book")
//}

func main() {
	book := Book{}
	book.Sell()
}
