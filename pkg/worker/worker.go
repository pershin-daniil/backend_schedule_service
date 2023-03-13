package worker

import (
	"context"
	"fmt"
	"time"

	"github.com/pershin-daniil/TimeSlots/pkg/notifier"

	"github.com/pershin-daniil/TimeSlots/pkg/models"
	"github.com/sirupsen/logrus"
)

type workerCode int

const (
	stop workerCode = -1
	run  workerCode = 1
)

type Store interface {
	UsersWithMeetings(ctx context.Context) ([]models.UserNotify, error)
	SwitchNotificationStatus(ctx context.Context, meetingID int) error
}

type Worker struct {
	log      *logrus.Logger
	store    Store
	notifier *notifier.Notifier
}

func New(log *logrus.Logger, store Store, notifier *notifier.Notifier) *Worker {
	return &Worker{
		log:      log,
		store:    store,
		notifier: notifier,
	}
}

func (w *Worker) SendNotificationBeforeTraining(ctx context.Context) error {
	worker := run
	go func() {
		<-ctx.Done()
		worker = stop
	}()
	for {
		usersToNotify, err := w.store.UsersWithMeetings(ctx)
		if err != nil {
			return fmt.Errorf("worker send notification faild: %w", err)
		}
		for _, user := range usersToNotify {
			if user.Notified {
				continue
			}
			msg := fmt.Sprintf("У вас тренировка в %s", user.StartAt.String())
			if err = w.notifier.NotifyTelegram(ctx, msg, user); err != nil {
				return fmt.Errorf("worker send notification faild: %w", err)
			}
			if err = w.store.SwitchNotificationStatus(ctx, user.MeetingID); err != nil {
				return fmt.Errorf("worker send notification faild: %w", err)
			}
		}
		time.Sleep(5 * time.Second)
		if worker == stop {
			break
		}
	}
	return nil
}
