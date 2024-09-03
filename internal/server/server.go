package server

import (
	"fmt"
	"net/url"

	"github.com/FollowLille/metrics/internal/config"
)

type Server struct {
	Address string
	Port    int64
}

func NewServer() *Server {
	return &Server{
		Address: config.Address,
		Port:    config.Port,
	}
}

func (c *Server) ChangeAddress(address string) error {
	_, err := url.ParseRequestURI(address)
	if err != nil {
		return err
	}
	c.Address = address
	return nil
}

func (c *Server) ChangePort(port int64) error {
	if port < 1024 || port > 65535 {
		return fmt.Errorf("некорректный порт: %d", port)
	}
	c.Port = port
	return nil
}
