package gorillahttp

import (
	"net/http"
)

// Routes adds routes and handlers to the router.
func (s *Service) Routes() {
	s.router.NotFoundHandler = http.HandlerFunc(NotFoundH)
	s.router.MethodNotAllowedHandler = http.HandlerFunc(MethodNotAllowedH)

	jsonAPI := s.router.PathPrefix("/api").Subrouter()

	// ts3 service routes.
	ts3v1 := jsonAPI.PathPrefix("/ts3/v1").Subrouter()
	ts3v1.HandleFunc("/createregisterrecord", s.CreateRegisterRecordH)
}
