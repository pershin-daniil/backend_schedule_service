package service

import (
	"context"
	"fmt"

	"github.com/pershin-daniil/TimeSlots/pkg/models"
	"github.com/sirupsen/logrus"
)

type Notifier interface {
	Notify(ctx context.Context, message string, userID int) error
}

type Store interface {
	GetUsers(ctx context.Context) ([]models.User, error)
	CreateUser(ctx context.Context, user models.User) (models.User, error)
	GetUser(ctx context.Context, id int) (models.User, error)
	UpdateUser(ctx context.Context, id int, user models.User) (models.User, error)
	DeleteUser(ctx context.Context, id int) (models.User, error)
	TruncateTable(ctx context.Context, table string) error
}

type ScheduleService struct {
	log      *logrus.Entry
	store    Store
	notifier Notifier
}

func NewScheduleService(log *logrus.Logger, store Store, notifier Notifier) *ScheduleService {
	s := ScheduleService{
		log:      log.WithField("component", "service"),
		store:    store,
		notifier: notifier,
	}
	return &s
}

func (s *ScheduleService) GetUsers(ctx context.Context) ([]models.User, error) {
	users, err := s.store.GetUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("err getting user from store: %w", err)
	}
	return users, nil
}

func (s *ScheduleService) CreateUser(ctx context.Context, user models.User) (models.User, error) {
	user, err := s.store.CreateUser(ctx, user)
	if err != nil {
		return models.User{}, fmt.Errorf("err creating user: %w", err)
	}
	if err = s.notifier.Notify(ctx, "user created", user.ID); err != nil {
		s.log.Errorf("err notifying user: %v", err)
	}
	return user, nil
}

func (s *ScheduleService) GetUser(ctx context.Context, id int) (models.User, error) {
	return s.store.GetUser(ctx, id)
}

func (s *ScheduleService) UpdateUser(ctx context.Context, id int, user models.User) (models.User, error) {
	return s.store.UpdateUser(ctx, id, user)
}

func (s *ScheduleService) DeleteUser(ctx context.Context, id int) (models.User, error) {
	return s.store.DeleteUser(ctx, id)
}

func (s *ScheduleService) Notify(ctx context.Context, message string, userID int) error {
	return s.notifier.Notify(ctx, message, userID)
}
