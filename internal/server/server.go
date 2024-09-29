package server

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"

	"github.com/FollowLille/metrics/internal/config"
	"github.com/FollowLille/metrics/internal/storage"
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
		return fmt.Errorf("invalid port: %d", port)
	}
	c.Port = port
	return nil
}

func (c *Server) SaveMetricsToFile(s *storage.MemStorage, file *os.File) error {
	metrics := s.GetAllMetrics()
	data, err := json.Marshal(metrics)
	if err != nil {
		return fmt.Errorf("can't marshal metrics: %s", err)
	}
	_, err = file.Write(data)
	if err != nil {
		return fmt.Errorf("can't write metrics to file: %s", err)
	}
	_, err = file.WriteString("\n")
	if err != nil {
		return fmt.Errorf("can't write metrics to file: %s", err)
	}
	return nil
}
