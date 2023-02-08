package main

import "fmt"

type Service struct {
	Options Options
}

type Options struct {
	A string
	B string
	C string
	D string
	E string
}

type Option func(*Options)

func SetA(a string) Option {
	return func(opts *Options) {
		opts.A = a
	}
}

func SetB(b string) Option {
	return func(opts *Options) {
		opts.B = b
	}
}

func NewService(opts ...Option) *Service {
	options := Options{}
	for _, o := range opts {
		o(&options)
	}
	return &Service{
		Options: options,
	}
}

func main() {
	service := NewService(SetA("a"), SetB("b"))
	fmt.Println(service.Options)
}
