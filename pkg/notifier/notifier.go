package notifier

import (
	"context"

	"github.com/sirupsen/logrus"
)

type DummyNotifier struct {
	log *logrus.Entry
}

func New(log *logrus.Logger) *DummyNotifier {
	return &DummyNotifier{
		log: log.WithField("component", "notifier"),
	}
}

func (n *DummyNotifier) Notify(_ context.Context, message string, user interface{}) error {
	n.log.Infof("notifying user %d: %s", user, message)
	return nil
}
