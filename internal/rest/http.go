package rest

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	_ "embed"
	"encoding/pem"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/sirupsen/logrus"
)

type Server struct {
	log       *logrus.Entry
	address   string
	version   string
	app       App
	server    *http.Server
	publicKey *rsa.PublicKey
}

//go:embed private_rsa.pub
var publicSigningKey []byte

func NewServer(log *logrus.Logger, app App, address, version string) *Server {
	s := Server{
		log:       log.WithField("component", "rest"),
		address:   address,
		version:   version,
		app:       app,
		publicKey: mustGetPublicKey(publicSigningKey),
	}
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Get("/version", s.versionHandler)
	r.Route("/api", func(r chi.Router) {
		r.Use(middleware.RequestLogger(&middleware.DefaultLogFormatter{Logger: s.log, NoColor: true}))
		r.Route("/v1", func(r chi.Router) {
			r.Post("/login", s.loginHandler)
			r.Post("/users", s.createUserHandler)
			r.Group(func(r chi.Router) {
				r.Use(s.jwtAuth)
				r.Get("/users", s.getUsersHandler)
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
	})
	s.server = &http.Server{
		Addr:              s.address,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
	}
	return &s
}

func (s *Server) Run(ctx context.Context) error {
	go func() {
		<-ctx.Done()
		gfCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := s.server.Shutdown(gfCtx); err != nil {
			s.log.Warnf("err shutting down properly")
		}
	}()
	s.log.Infof("starting server on %s", s.address)
	if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func mustGetPublicKey(keyBytes []byte) *rsa.PublicKey {
	if len(keyBytes) == 0 {
		panic("file public.pub is missing or invalid")
	}
	block, _ := pem.Decode(keyBytes)
	if block == nil {
		panic("unable to decode public key to blocks")
	}
	key, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		panic(err)
	}
	return key.(*rsa.PublicKey)
}
