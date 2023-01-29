package service

import (
	"context"
	"fmt"
	"github.com/pershin-daniil/TimeSlots/pkg/models"
	"github.com/sirupsen/logrus"
)

type Notifier interface {
	Notify(ctx context.Context, message string, user interface{}) error
}

type Store interface {
	GetUsers(ctx context.Context) ([]models.User, error)
	CreateUser(ctx context.Context, user models.User) (models.User, error)
	GetUser(ctx context.Context, id int) (models.User, error)
	UpdateUser(ctx context.Context, id int, user models.User) (models.User, error)
	DeleteUser(ctx context.Context, id int) (models.User, error)
	ResetTables(ctx context.Context, table []string) error
	GetMeetings(ctx context.Context) ([]models.Meeting, error)
	CreateMeeting(ctx context.Context, meeting models.Meeting) (models.Meeting, error)
	GetMeeting(ctx context.Context, id int) (models.Meeting, error)
	UpdateMeeting(ctx context.Context, id int, data models.Meeting) (models.Meeting, error)
	DeleteMeeting(ctx context.Context, id int) (models.Meeting, error)
}

type ScheduleService struct {
	log      *logrus.Entry
	store    Store
	notifier Notifier
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
		return nil, fmt.Errorf("err getting users from store: %w", err)
	}
	return users, nil
}

func (s *ScheduleService) GetUser(ctx context.Context, id int) (models.User, error) {
	user, err := s.store.GetUser(ctx, id)
	if err != nil {
		return models.User{}, fmt.Errorf("err getting user (id %d) from store: %w", id, err)
	}
	if err = s.notifier.Notify(ctx, "user", user); err != nil {
		s.log.Errorf("err notifying user: %v", err)
	}
	return user, nil
}

func (s *ScheduleService) UpdateUser(ctx context.Context, id int, user models.User) (models.User, error) {
	updatedUser, err := s.store.UpdateUser(ctx, id, user)
	if err != nil {
		return models.User{}, fmt.Errorf("err updating user (id %d) from store: %w", id, err)
	}
	if err = s.notifier.Notify(ctx, "user updated", updatedUser); err != nil {
		s.log.Errorf("err notifying user: %v", err)
	}
	return updatedUser, nil
}

func (s *ScheduleService) DeleteUser(ctx context.Context, id int) (models.User, error) {
	deletedUser, err := s.store.DeleteUser(ctx, id)
	if err != nil {
		return models.User{}, fmt.Errorf("err deleting user (id %d) from store: %w", id, err)
	}
	if err = s.notifier.Notify(ctx, "user deleted", deletedUser); err != nil {
		s.log.Errorf("err notifying user: %v", err)
	}
	return deletedUser, nil
}

func (s *ScheduleService) CreateMeeting(ctx context.Context, meeting models.Meeting) (models.Meeting, error) {
	createdMeeting, err := s.store.CreateMeeting(ctx, meeting)
	if err != nil {
		return models.Meeting{}, fmt.Errorf("err creating meeting: %w", err)
	}
	if err = s.notifier.Notify(ctx, "meeting created", createdMeeting); err != nil {
		s.log.Errorf("err notifying user: %v", err)
	}
	return createdMeeting, nil
}

func (s *ScheduleService) GetMeetings(ctx context.Context) ([]models.Meeting, error) {
	meetings, err := s.store.GetMeetings(ctx)
	if err != nil {
		return nil, fmt.Errorf("err getting meetings: %w", err)
	}
	return meetings, nil
}

func (s *ScheduleService) GetMeeting(ctx context.Context, id int) (models.Meeting, error) {
	meeting, err := s.store.GetMeeting(ctx, id)
	if err != nil {
		return models.Meeting{}, fmt.Errorf("err getting meeting (id %d) from store: %w", id, err)
	}
	if err = s.notifier.Notify(ctx, "meeting", meeting); err != nil {
		s.log.Errorf("err notifying user: %v", err)
	}
	return meeting, nil
}

func (s *ScheduleService) UpdateMeeting(ctx context.Context, id int, meeting models.Meeting) (models.Meeting, error) {
	updatedMeeting, err := s.store.UpdateMeeting(ctx, id, meeting)
	if err != nil {
		return models.Meeting{}, fmt.Errorf("err updating meeting (id %d) from store: %w", id, err)
	}
	if err = s.notifier.Notify(ctx, "meeting updated", updatedMeeting); err != nil {
		s.log.Errorf("err notifying user: %v", err)
	}
	return updatedMeeting, nil
}

func (s *ScheduleService) DeleteMeeting(ctx context.Context, id int) (models.Meeting, error) {
	deletedMeeting, err := s.store.DeleteMeeting(ctx, id)
	if err != nil {
		return models.Meeting{}, fmt.Errorf("err deleting meeting (id %d) from store: %w", id, err)
	}
	if err = s.notifier.Notify(ctx, "meeting deleted", deletedMeeting); err != nil {
		s.log.Errorf("err notifying user: %v", err)
	}
	return deletedMeeting, nil
}

func (s *ScheduleService) Notify(ctx context.Context, message string, user interface{}) error {
	return s.notifier.Notify(ctx, message, user)
}
