package rest

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/pershin-daniil/TimeSlots/pkg/store"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/pershin-daniil/TimeSlots/pkg/models"
	"github.com/sirupsen/logrus"
)

type App interface {
	GetUsers(ctx context.Context) ([]models.User, error)
	CreateUser(ctx context.Context, user models.User) (models.User, error)
	ReadUser(ctx context.Context, id string) ([]models.User, error)
	UpdateUser(ctx context.Context, id string, user models.User) ([]models.User, error)
	DeleteUser(ctx context.Context, id string) ([]models.User, error)
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
	r.Get("/version", h.versionHandler)
	r.Route("/api", func(r chi.Router) {
		r.Route("/v1", func(r chi.Router) {
			r.Get("/users", h.getUsersHandler)
			r.Post("/users", h.createUserHandler)
			r.Get("/users/{id}", h.searchUserHandler)
			r.Patch("/users/{id}", h.updateUserHandler)
			r.Delete("/users/{id}", h.deleteUserHandler)
		})
	})

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
		http.Error(w, "internal sever error", http.StatusInternalServerError)
		return
	}
	if err = json.NewEncoder(w).Encode(users); err != nil {
		h.log.Warnf("err during econding users: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) createUserHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	createdUser, err := h.app.CreateUser(ctx, user)
	if err != nil {
		h.log.Warnf("err during creating user: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(createdUser); err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) searchUserHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := chi.URLParam(r, "id")
	user, err := h.app.ReadUser(ctx, id)
	if err != nil {
		h.log.Warnf("err during getting users: %v", err)
		http.Error(w, "internal sever error", http.StatusInternalServerError)
		return
	}
	if err = json.NewEncoder(w).Encode(user); err != nil {
		h.log.Warnf("err during econding users: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) updateUserHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := chi.URLParam(r, "id")
	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	updatedUser, err := h.app.UpdateUser(ctx, id, user)
	if err != nil {
		h.log.Warnf("err during creating user: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(updatedUser); err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) deleteUserHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := chi.URLParam(r, "id")
	deletedUser, err := h.app.DeleteUser(ctx, id)
	if err != nil {
		h.log.Warnf("err during creating user: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(deletedUser); err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
}
