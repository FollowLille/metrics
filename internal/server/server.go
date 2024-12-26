// Package server отвечает за запуск сервера
package server

import (
	"fmt"
	"net/url"

	"github.com/FollowLille/metrics/internal/config"
)

// Server хранит адрес и порт сервера
type Server struct {
	Address string
	Port    int64
}

// NewServer создает новый Server
// и возвращает его в виде *Server
//
// Возвращаемое значение:
//   - *Server
func NewServer() *Server {
	return &Server{
		Address: config.Address,
		Port:    config.Port,
	}
}

// ChangeAddress изменяет адрес сервера
//
// Параметры:
//   - address - адрес сервера
//
// Возвращаемое значение:
//   - error
func (c *Server) ChangeAddress(address string) error {
	_, err := url.ParseRequestURI(address)
	if err != nil {
		return err
	}
	c.Address = address
	return nil
}

// ChangePort изменяет порт сервера
//
// Параметры:
//   - port - порт сервера
//
// Возвращаемое значение:
//   - error
func (c *Server) ChangePort(port int64) error {
	if port < 1024 || port > 65535 {
		return fmt.Errorf("invalid port: %d", port)
	}
	c.Port = port
	return nil
}
