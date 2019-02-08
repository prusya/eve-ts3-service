package gorillahttp

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"

	"github.com/prusya/eve-ts3-service/pkg/system"
)

var (
	serviceName = "gorillahttp"
)

// Service implements http.Service interface backed by gorilla toolkit.
type Service struct {
	system *system.System
	router *mux.Router
	server *http.Server
}

// New creates a new Service and prepares it to Start.
func New(system *system.System) *Service {
	r := mux.NewRouter()

	s := Service{
		system: system,
		router: r,
		server: &http.Server{
			Addr:         system.Config.WebServerAddress,
			Handler:      r,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  20 * time.Second,
		},
	}
	s.Routes()

	s.system.HTTP = &s

	return &s
}

// Start starts the Service.
func (s *Service) Start() {
	go func() {
		err := s.server.ListenAndServe()
		if err != nil {
			if err != http.ErrServerClosed {
				log.Println(err)
			}
			s.system.SigChan <- os.Interrupt
		}
	}()
}

// Stop stops the Service.
func (s *Service) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	s.server.Shutdown(ctx)
}

// ServiceName return this Service's name.
func (s *Service) ServiceName() string {
	return "HTTP"
}
