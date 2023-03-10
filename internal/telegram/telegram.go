package telegram

import (
	"context"
	"fmt"
	"time"

	"github.com/pershin-daniil/TimeSlots/pkg/models"

	"github.com/sirupsen/logrus"
	tele "gopkg.in/telebot.v3"
)

type Telegram struct {
	log *logrus.Entry
	bot *tele.Bot
	app App
}

type Notifier struct {
	log *logrus.Entry
	bot *tele.Bot
}

type App interface {
	CreateUser(ctx context.Context, user models.UserRequest) (models.User, error)
}

func NewNotifier(log *logrus.Logger, bot *tele.Bot) *Notifier {
	return &Notifier{
		log: log.WithField("component", "notifier"),
		bot: bot,
	}
}

func New(log *logrus.Logger, bot *tele.Bot, app App) (*Telegram, error) {
	t := Telegram{
		log: log.WithField("component", "telegram"),
		bot: bot,
		app: app,
	}
	t.initButtons()
	t.initHandlers()
	return &t, nil
}

func NewBot(token string) (*tele.Bot, error) {
	config := tele.Settings{
		Token:  token,
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	}
	b, err := tele.NewBot(config)
	if err != nil {
		return nil, fmt.Errorf("new bot faild: %w", err)
	}
	return b, nil
}

func (t *Notifier) Notify(ctx context.Context, msg string, user interface{}) error {
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
