package telegram

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	tele "gopkg.in/telebot.v3"
)

type Telegram struct {
	log *logrus.Entry
	bot *tele.Bot
}

func New(log *logrus.Logger, token string) (*Telegram, error) {
	config := tele.Settings{
		Token:  token,
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	}
	b, err := tele.NewBot(config)
	if err != nil {
		return nil, fmt.Errorf("new bot faild: %w", err)
	}
	t := Telegram{
		log: logrus.WithField("component", "telegram"),
		bot: b,
	}
	t.initButtons()
	t.initHandlers()
	return &t, nil
}

func (t *Telegram) Notify(ctx context.Context, msg string, user interface{}) error {
	t.log.Infof("Notification: %v %v", msg, user)
	return nil
}

func (t *Telegram) Run(ctx context.Context) {
	go func() {
		<-ctx.Done()
		t.bot.Stop()
	}()
	t.log.Infof("Starting telegram bot as %v", t.bot.Me.Username)
	t.bot.Start()
}
