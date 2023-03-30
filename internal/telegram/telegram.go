package telegram

import (
	"context"
	"fmt"
	"github.com/pershin-daniil/TimeSlots/internal/calendar"
	"time"

	"github.com/pershin-daniil/TimeSlots/pkg/models"

	"github.com/sirupsen/logrus"
	tele "gopkg.in/telebot.v3"
)

type Telegram struct {
	log *logrus.Entry
	bot *tele.Bot
	app App
	cal Calendar
}

type App interface {
	CreateUser(ctx context.Context, user models.UserRequest) (models.User, error)
}

type Calendar interface {
	Events() []models.Event
}

func New(log *logrus.Logger, bot *tele.Bot, app App, cal *calendar.Calendar) (*Telegram, error) {
	t := Telegram{
		log: log.WithField("module", "telegram"),
		bot: bot,
		app: app,
		cal: cal,
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

func (t *Telegram) Run(ctx context.Context) {
	go func() {
		<-ctx.Done()
		t.bot.Stop()
	}()
	t.log.Infof("Starting telegram bot as %v", t.bot.Me.Username)
	t.bot.Start()
}
