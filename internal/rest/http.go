package rest

import (
	"context"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

type Server struct {
	log     *logrus.Entry
	address string
	version string
	app     App
	server  *http.Server
}

func NewServer(log *logrus.Logger, app App, address, version string) *Server {
	s := Server{
		log:     log.WithField("component", "rest"),
		address: address,
		version: version,
		app:     app,
	}
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Get("/version", s.versionHandler)
	r.Route("/api", func(r chi.Router) {
		r.Use(middleware.RequestLogger(&middleware.DefaultLogFormatter{Logger: s.log, NoColor: true}))
		r.Route("/v1", func(r chi.Router) {
			r.Get("/users", s.getUsersHandler)
			r.Post("/users", s.createUserHandler)
			r.Get("/users/{id}", s.getUserHandler)
			r.Patch("/users/{id}", s.updateUserHandler)
			r.Delete("/users/{id}", s.deleteUserHandler)
			r.Post("/meetings", s.createMeetingHandler)
			r.Get("/meetings", s.getMeetingsHandler)
			r.Get("/meetings/{id}", s.getMeetingHandler)
			r.Patch("/meetings/{id}", s.updateMeetingHandler)
			r.Delete("/meetings/{id}", s.deleteMeetingHandler)
		})
	})
	s.server = &http.Server{
		Addr:              s.address,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
	}
	return &s
}

func (s *Server) Run() error {
	s.log.Infof("starting server on %s", s.address)
	if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func (s *Server) Shutdown() {
	// TODO: ctx, proper error handling and timeout - later
	if err := s.server.Shutdown(context.Background()); err != nil {
		s.log.Warnf("err during shutting down server: %v", err)
	}
}
