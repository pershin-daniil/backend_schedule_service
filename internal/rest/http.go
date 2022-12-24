package rest

import (
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/sirupsen/logrus"
	"net/http"
	"time"
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
	_, err := fmt.Fprintf(w, "%s\n", h.version)
	if err != nil {
		h.log.Warnf("err during writing to connection: %v", err)
	}
}

func (h *Handler) timeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/timeNow" {
		http.Error(w, "404 not found.", http.StatusNotFound)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method is not supported.", http.StatusNotFound)
		return
	}
	h.stateTime = append(h.stateTime, time.Now())
	h.log.Debugf("User requested current time at %s", time.Now().Format(time.RFC3339))
	_, err := fmt.Fprintf(w, "%s\n", h.stateTime[len(h.stateTime)-1])
	if err != nil {
		h.log.Panic(err)
	}
}

func (h *Handler) timeHistoryHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/timeHistory" {
		http.Error(w, "404 not found.", http.StatusNotFound)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method is not supported.", http.StatusNotFound)
		return
	}
	h.log.Debugf("User requested time history at %s", time.Now().Format(time.RFC3339))
	_, err := fmt.Fprintf(w, "%v\n", h.stateTime)
	if err != nil {
		h.log.Panic(err)
	}
}

func (h *Handler) resetTimeHistoryHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/resetTimeHistory" {
		http.Error(w, "404 not found.", http.StatusNotFound)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method is not supported.", http.StatusNotFound)
		return
	}
	h.log.Debugf("User requested reset time history at %s", time.Now().Format(time.RFC3339))
	_, err := fmt.Fprintf(w, "%v\n", h.stateTime)
	if err != nil {
		h.log.Panic(err)
	}
	h.stateTime = h.stateTime[:0]
}
