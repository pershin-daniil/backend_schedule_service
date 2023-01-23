package notifier

import (
	"context"

	"github.com/sirupsen/logrus"
)

type DummyNotifier struct {
	log *logrus.Entry
}

func NewDummyNotifier(log *logrus.Logger) *DummyNotifier {
	return &DummyNotifier{
		log: log.WithField("component", "notifier"),
	}
}

func (n *DummyNotifier) Notify(_ context.Context, message string, userID int) error {
	n.log.Infof("notifying user %d: %s", userID, message)
	return nil
}
