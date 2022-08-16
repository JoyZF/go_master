package functional_options

import (
	"crypto/tls"
	"time"
)

type Config struct {
	Protocol string
	Timeout  time.Duration
	Maxconns int
	TLS      *tls.Config
}

type Server struct {
	Addr string
	Port int
	Conf *Config
}

type Option func(*Server)


func Protocol(p string) Option {
	return func(s *Server) {
		s.Conf.Protocol = p
	}
}

func Timeout(t time.Duration) Option {
	return func(s *Server) {
		s.Conf.Timeout = t
	}
}

func MaxConns(maxconns int) Option {
	return func(s *Server) {
		s.Conf.Maxconns = maxconns
	}
}

func TLS(tls *tls.Config) Option {
	return func(s *Server) {
		s.Conf.TLS = tls
	}
}


func NewServer(addr string, port int, options ...func(*Server)) (*Server, error) {

	srv := Server{
		Addr:     addr,
		Port:     port,
	}
	for _, option := range options {
		option(&srv)
	}
	//...
	return &srv, nil
}

func Init()  {
	_, _ = NewServer("127.0.0.1", 6379, Timeout(3*time.Second), Protocol("123")
}