package notifier

import (
	"context"
	"fmt"

	"github.com/pershin-daniil/TimeSlots/pkg/models"
	"github.com/sirupsen/logrus"
	tele "gopkg.in/telebot.v3"
)

type Notifier struct {
	log *logrus.Entry
	bot *tele.Bot
}

func New(log *logrus.Logger, bot *tele.Bot) *Notifier {
	return &Notifier{
		log: log.WithField("module", "notifier"),
		bot: bot,
	}
}

func (n *Notifier) NotifyTelegram(_ context.Context, msg string, data models.UserNotify) error {
	n.log.Infof("Notification: %v %v", msg, data)
	chat, err := n.bot.ChatByID(int64(data.UserID))
	if err != nil {
		return fmt.Errorf("notify telegram faild: %w", err)
	}
	if _, err = n.bot.Send(chat, fmt.Sprintf("%v %v", msg, data)); err != nil {
		return fmt.Errorf("notify telegram faild: %w", err)
	}
	return nil
}
