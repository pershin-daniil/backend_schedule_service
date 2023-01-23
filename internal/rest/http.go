package rest

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/pershin-daniil/TimeSlots/pkg/store"
	"net/http"
	"strconv"

	"github.com/pershin-daniil/TimeSlots/pkg/models"
	"github.com/sirupsen/logrus"
)

type App interface {
	GetUsers(ctx context.Context) ([]models.User, error)
	CreateUser(ctx context.Context, user models.User) (models.User, error)
	GetUser(ctx context.Context, id int) (models.User, error)
	UpdateUser(ctx context.Context, id int, user models.User) (models.User, error)
	DeleteUser(ctx context.Context, id int) (models.User, error)
}

type Handler struct {
	log     *logrus.Entry
	address string
	version string
	app     App
}

func NewHandler(log *logrus.Logger, app *store.Store, address, version string) *Handler {
	h := Handler{
		log:     log.WithField("component", "rest"),
		address: address,
		version: version,
		app:     app,
	}
	return &h
}

func (h *Handler) Run() error {
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Get("/version", h.versionHandler)
	r.Route("/api", func(r chi.Router) {
		r.Use(middleware.RequestLogger(&middleware.DefaultLogFormatter{Logger: h.log, NoColor: true}))
		r.Route("/v1", func(r chi.Router) {
			r.Get("/users", h.getUsersHandler)
			r.Post("/users", h.createUserHandler)
			r.Get("/users/{id}", h.getUserHandler)
			r.Patch("/users/{id}", h.updateUserHandler)
			r.Delete("/users/{id}", h.deleteUserHandler)
		})
	})

	h.log.Infof("starting server on %s", h.address)
	if err := http.ListenAndServe(h.address, r); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func (h *Handler) versionHandler(w http.ResponseWriter, _ *http.Request) {
	_, err := fmt.Fprintf(w, "%s\n", h.version)
	if err != nil {
		h.log.Warnf("err during writing to connection: %v", err)
	}
}

func (h *Handler) getUsersHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	users, err := h.app.GetUsers(ctx)
	if err != nil {
		h.log.Warnf("err during getting users: %v", err)
		h.writeResponse(w, http.StatusInternalServerError, err)
		return
	}
	h.writeResponse(w, http.StatusOK, users)
}

func (h *Handler) createUserHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		h.writeResponse(w, http.StatusBadRequest, err)
		return
	}
	createdUser, err := h.app.CreateUser(ctx, user)
	if err != nil {
		h.log.Warnf("err during creating user: %v", err)
		h.writeResponse(w, http.StatusInternalServerError, err)
		return
	}
	h.writeResponse(w, http.StatusCreated, createdUser)
}

func (h *Handler) getUserHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		h.writeResponse(w, http.StatusBadRequest, err)
		return
	}
	user, err := h.app.GetUser(ctx, id)
	switch {
	case errors.Is(err, store.ErrUserNotFound):
		h.writeResponse(w, http.StatusNotFound, err)
		return
	case err != nil:
		h.log.Warnf("err during getting users: %v", err)
		h.writeResponse(w, http.StatusInternalServerError, err)
		return
	}
	h.writeResponse(w, http.StatusOK, user)
}

func (h *Handler) updateUserHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		h.writeResponse(w, http.StatusBadRequest, err)
		return
	}
	var newData models.User
	if err := json.NewDecoder(r.Body).Decode(&newData); err != nil {
		h.writeResponse(w, http.StatusBadRequest, err)
		return
	}
	updatedUser, err := h.app.UpdateUser(ctx, id, newData)
	switch {
	case errors.Is(err, store.ErrUserNotFound):
		h.writeResponse(w, http.StatusNotFound, err)
		return
	case err != nil:
		h.log.Warnf("err during updating users: %v", err)
		h.writeResponse(w, http.StatusInternalServerError, err)
		return
	}
	h.writeResponse(w, http.StatusOK, updatedUser)
}

func (h *Handler) deleteUserHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		h.writeResponse(w, http.StatusBadRequest, err)
		return
	}
	deletedUser, err := h.app.DeleteUser(ctx, id)
	switch {
	case errors.Is(err, store.ErrUserNotFound):
		h.writeResponse(w, http.StatusNotFound, err)
		return
	case err != nil:
		h.log.Warnf("err during deleting users: %v", err)
		h.writeResponse(w, http.StatusInternalServerError, err)
		return
	}
	h.writeResponse(w, http.StatusOK, deletedUser)
}

func (h *Handler) writeResponse(w http.ResponseWriter, status int, data interface{}) {
	w.WriteHeader(status)
	w.Header().Set("Content-Type", "application/json")
	if x, ok := data.(error); ok {
		if err := json.NewEncoder(w).Encode(ErrorResponse{Err: x.Error()}); err != nil {
			h.log.Warnf("err during encoding error: %v", err)
		}
		return
	}
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.log.Warnf("err during encoding responce: %v", err)
	}
}

type ErrorResponse struct {
	Err string `json:"error"`
}
