package telegram

import (
	"context"
	"github.com/sirupsen/logrus"
	tele "gopkg.in/telebot.v3"
	"time"
)

type Telegram struct {
	log *logrus.Entry
	bot *tele.Bot
}

func NewTelegram(log *logrus.Logger, token string) (*Telegram, error) {
	config := tele.Settings{
		Token:  token,
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	}
	b, err := tele.NewBot(config)
	if err != nil {
		return nil, err
	}
	t := Telegram{
		log: logrus.WithField("component", "telegram"),
		bot: b,
	}
	t.initHandlers()
	return &t, nil
}

func (t *Telegram) Run(ctx context.Context) {
	go func() {
		<-ctx.Done()
		t.bot.Stop()
	}()
	t.log.Infof("starting telegram bot as %v", t.bot.Me.Username)
	t.bot.Start()
}

func (t *Telegram) Notify(ctx context.Context, message string, user interface{}) error {
	t.log.Infof("Notification: %v %v", message, user)
	return nil
}

func (t *Telegram) initHandlers() {
	t.bot.Handle(CommandHello, t.helloHandler)
}

func (t *Telegram) helloHandler(c tele.Context) error {
	return c.Send("Hello!")
}
