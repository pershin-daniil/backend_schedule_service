package rest

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	log       *logrus.Entry
	address   string
	version   string
	stateTime []time.Time
}

func NewHandler(log *logrus.Logger, address, version string) *Handler {
	h := Handler{
		log:     log.WithField("component", "rest"),
		address: address,
		version: version,
	}
	return &h
}

func (h *Handler) Run() error {
	r := chi.NewRouter()
	r.Get("/version", h.versionHandler)
	r.Route("/api", func(r chi.Router) {
		r.Route("/v1", func(r chi.Router) {
			r.Get("/timeNow", h.timeHandler)
			r.Get("/timeHistory", h.timeHistoryHandler)
			r.Get("/resetTimeHistory", h.resetTimeHistoryHandler)
		})
	})

	if err := http.ListenAndServe(h.address, r); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func (h *Handler) versionHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/version" {
		http.Error(w, "404 not found.", http.StatusNotFound)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "Method is not allowed", http.StatusMethodNotAllowed)
		return
	}
	_, err := fmt.Fprintf(w, "%s\n", h.version)
	if err != nil {
		h.log.Warnf("err during writing to connection: %v", err)
	}
}

func (h *Handler) timeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/api/v1/timeNow" {
		http.Error(w, "404 not found.", http.StatusNotFound)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "Method is not allowed.", http.StatusMethodNotAllowed)
		return
	}
	h.stateTime = append(h.stateTime, time.Now())
	h.log.Debug("User requested current time")
	_, err := fmt.Fprintf(w, "%s\n", h.stateTime[len(h.stateTime)-1])
	if err != nil {
		h.log.Panic(err)
	}
}

func (h *Handler) timeHistoryHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/api/v1/timeHistory" {
		http.Error(w, "404 not found.", http.StatusNotFound)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "Method is not allowed.", http.StatusMethodNotAllowed)
		return
	}
	h.log.Debug("User requested time history")
	_, err := fmt.Fprintf(w, "%v\n", h.stateTime)
	if err != nil {
		h.log.Panic(err)
	}
}

func (h *Handler) resetTimeHistoryHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/api/v1/resetTimeHistory" {
		http.Error(w, "404 not found.", http.StatusNotFound)
	}
	if r.Method != http.MethodGet {
		http.Error(w, "Method is not allowed.", http.StatusMethodNotAllowed)
		return
	}
	h.log.Debug("User requested to reset time history")
	_, err := fmt.Fprintf(w, "%v\n", h.stateTime)
	if err != nil {
		h.log.Panic(err)
	}
	h.stateTime = h.stateTime[:0]
}
