package rest

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/pershin-daniil/TimeSlots/pkg/models"
	"github.com/pershin-daniil/TimeSlots/pkg/pgstore"
	"github.com/pershin-daniil/TimeSlots/pkg/service"
	"net/http"
	"strconv"
)

type App interface {
	GetUsers(ctx context.Context) ([]models.User, error)
	CreateUser(ctx context.Context, user models.User) (models.User, error)
	GetUser(ctx context.Context, id int) (models.User, error)
	UpdateUser(ctx context.Context, id int, user models.User) (models.User, error)
	DeleteUser(ctx context.Context, id int) (models.User, error)
	GetMeetings(ctx context.Context) ([]models.Meeting, error)
	CreateMeeting(ctx context.Context, meeting models.Meeting) (models.Meeting, error)
	GetMeeting(ctx context.Context, id int) (models.Meeting, error)
	UpdateMeeting(ctx context.Context, id int, meeting models.Meeting) (models.Meeting, error)
	DeleteMeeting(ctx context.Context, id int) (models.Meeting, error)
	service.Notifier
}

func (s *Server) versionHandler(w http.ResponseWriter, _ *http.Request) {
	_, err := fmt.Fprintf(w, "%s\n", s.version)
	if err != nil {
		s.log.Warnf("err during writing to connection: %v", err)
	}
}

func (s *Server) getUsersHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	users, err := s.app.GetUsers(ctx)
	if err != nil {
		s.log.Warnf("err during getting users: %v", err)
		s.writeResponse(w, http.StatusInternalServerError, err)
		return
	}
	s.writeResponse(w, http.StatusOK, users)
}

func (s *Server) createUserHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		s.writeResponse(w, http.StatusBadRequest, err)
		return
	}
	createdUser, err := s.app.CreateUser(ctx, user)
	if err != nil {
		s.log.Warnf("err during creating user: %v", err)
		s.writeResponse(w, http.StatusInternalServerError, err)
		return
	}
	s.writeResponse(w, http.StatusCreated, createdUser)
}

func (s *Server) getUserHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		s.writeResponse(w, http.StatusBadRequest, err)
		return
	}
	user, err := s.app.GetUser(ctx, id)
	switch {
	case errors.Is(err, pgstore.ErrUserNotFound):
		s.writeResponse(w, http.StatusNotFound, err)
		return
	case err != nil:
		s.log.Warnf("err during getting users: %v", err)
		s.writeResponse(w, http.StatusInternalServerError, err)
		return
	}
	s.writeResponse(w, http.StatusOK, user)
}

func (s *Server) updateUserHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		s.writeResponse(w, http.StatusBadRequest, err)
		return
	}
	var newData models.User
	if err := json.NewDecoder(r.Body).Decode(&newData); err != nil {
		s.writeResponse(w, http.StatusBadRequest, err)
		return
	}
	updatedUser, err := s.app.UpdateUser(ctx, id, newData)
	switch {
	case errors.Is(err, pgstore.ErrUserNotFound):
		s.writeResponse(w, http.StatusNotFound, err)
		return
	case err != nil:
		s.log.Warnf("err during updating users: %v", err)
		s.writeResponse(w, http.StatusInternalServerError, err)
		return
	}
	s.writeResponse(w, http.StatusOK, updatedUser)
}

func (s *Server) deleteUserHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		s.writeResponse(w, http.StatusBadRequest, err)
		return
	}
	deletedUser, err := s.app.DeleteUser(ctx, id)
	switch {
	case errors.Is(err, pgstore.ErrUserNotFound):
		s.writeResponse(w, http.StatusNotFound, err)
		return
	case err != nil:
		s.log.Warnf("err during deleting users: %v", err)
		s.writeResponse(w, http.StatusInternalServerError, err)
		return
	}
	s.writeResponse(w, http.StatusOK, deletedUser)
}

func (s *Server) createMeetingHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var meeting models.Meeting
	if err := json.NewDecoder(r.Body).Decode(&meeting); err != nil {
		s.writeResponse(w, http.StatusBadRequest, err)
		return
	}
	createdMeeting, err := s.app.CreateMeeting(ctx, meeting)
	if err != nil {
		s.log.Warnf("err during creating meeeting: %v", err)
		s.writeResponse(w, http.StatusInternalServerError, err)
		return
	}
	s.writeResponse(w, http.StatusCreated, createdMeeting)
}

func (s *Server) getMeetingsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	meetings, err := s.app.GetMeetings(ctx)
	if err != nil {
		s.log.Warnf("err during getting meetings: %v", err)
		s.writeResponse(w, http.StatusInternalServerError, err)
		return
	}
	s.writeResponse(w, http.StatusOK, meetings)
}

func (s *Server) getMeetingHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		s.writeResponse(w, http.StatusBadRequest, err)
		return
	}
	meeting, err := s.app.GetMeeting(ctx, id)
	switch {
	case errors.Is(err, pgstore.ErrMeetingNotFound):
		s.writeResponse(w, http.StatusNotFound, err)
	case err != nil:
		s.log.Warnf("err during getting meeting: %v", err)
		s.writeResponse(w, http.StatusInternalServerError, err)
		return
	}
	s.writeResponse(w, http.StatusOK, meeting)
}

func (s *Server) updateMeetingHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		s.writeResponse(w, http.StatusBadRequest, err)
		return
	}
	var newData models.Meeting
	if err := json.NewDecoder(r.Body).Decode(&newData); err != nil {
		s.writeResponse(w, http.StatusBadRequest, err)
		return
	}
	updatedMeeting, err := s.app.UpdateMeeting(ctx, id, newData)
	switch {
	case errors.Is(err, pgstore.ErrMeetingNotFound):
		s.writeResponse(w, http.StatusNotFound, err)
	case err != nil:
		s.log.Warnf("err during updating meeting: %v", err)
		s.writeResponse(w, http.StatusInternalServerError, err)
		return
	}
	s.writeResponse(w, http.StatusOK, updatedMeeting)
}

func (s *Server) deleteMeetingHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		s.writeResponse(w, http.StatusBadRequest, err)
		return
	}
	deletedMeeting, err := s.app.DeleteMeeting(ctx, id)
	switch {
	case errors.Is(err, pgstore.ErrMeetingNotFound):
		s.writeResponse(w, http.StatusNotFound, err)
	case err != nil:
		s.log.Warnf("err during deleting meeting: %v", err)
		s.writeResponse(w, http.StatusInternalServerError, err)
		return
	}
	s.writeResponse(w, http.StatusOK, deletedMeeting)
}

func (s *Server) writeResponse(w http.ResponseWriter, status int, data interface{}) {
	w.WriteHeader(status)
	w.Header().Set("Content-Type", "application/json")
	if x, ok := data.(error); ok {
		if err := json.NewEncoder(w).Encode(ErrorResponse{Err: x.Error()}); err != nil {
			s.log.Warnf("err during encoding error: %v", err)
		}
		return
	}
	if err := json.NewEncoder(w).Encode(data); err != nil {
		s.log.Warnf("err during encoding responce: %v", err)
	}
}

type ErrorResponse struct {
	Err string `json:"error"`
}
