package api

import (
	"net/http"
	"net/url"
)

func NewClient(endpoint string) (*Server, error) {

	ep, err := url.Parse(endpoint)

	if err != nil {
		return nil, err
	}

	return &Server{&client{endpoint: ep, httpConn: http.DefaultClient}}, nil
}

type Server struct {
	c *client
}

func (s *Server) Ping() error {
	path := "/_ping"
	body, status, err := s.c.do("GET", path, nil)
	if err != nil {
		return err
	}
	if status != http.StatusOK {
		return newError(status, body)
	}
	return nil
}
