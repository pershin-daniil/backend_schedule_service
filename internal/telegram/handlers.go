package telegram

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/pershin-daniil/TimeSlots/pkg/models"

	tele "gopkg.in/telebot.v3"
)

var coachCode = "TRAIN"

func (t *Telegram) initHandlers() {
	t.bot.Handle(cmdStart, t.startHandler)
	t.bot.Handle(&registrationBtn, t.registrationHandler)
	t.bot.Handle(&availableMeetingsBtn, t.scheduleHandler)
	t.bot.Handle(&myMeetingBtn, t.meetingsHandler)
	t.bot.Handle(&notificationBtn, t.notifyHandler)
	t.bot.Handle(&cancelMeetingBtn, t.cancelMeetingHandler)
	t.bot.Handle(tele.OnText, t.textHandler)
}

func (t *Telegram) textHandler(ctx tele.Context) error {
	// TODO: проверка на существование пользователя
	if ctx.Text() == coachCode {
		if err := ctx.Send("У тебя самый лучший тренер, го посмотрим его расписание", availMeetings); err != nil {
			return fmt.Errorf("tg send message faild: %w", err)
		}
	} else {
		if err := ctx.Send("Не понял, давай сначала"); err != nil {
			return fmt.Errorf("tg send message faild: %w", err)
		}
	}
	return nil
}

func (t *Telegram) startHandler(ctx tele.Context) error {
	msg := `Вступительная речь
Предложение продолжить`
	// TODO: проверка на существование пользователя
	parseUserRequest(ctx)
	return ctx.Send(msg, registration)
}

func parseUserRequest(ctx tele.Context) models.UserRequest {
	password := uuid.New().String()
	role := models.RoleClient

	return models.UserRequest{
		LastName:  &ctx.Sender().LastName,
		FirstName: &ctx.Sender().FirstName,
		Role:      &role,
		Password:  &password,
	}
}

func (t *Telegram) registrationHandler(ctx tele.Context) error {
	msg := "Введите код тренера"
	return ctx.Edit(msg)
}

func (t *Telegram) scheduleHandler(ctx tele.Context) error {
	// TODO: кейс записи на тренировку
	msg := "Здесь доступное расписание тренера, на которое можно записаться."
	return ctx.Edit(msg, showMeetings)
}

func (t *Telegram) meetingsHandler(ctx tele.Context) error {
	// TODO: отображение расписания
	msg := "Моё расписание"
	return ctx.Edit(msg, settings)
}

func (t *Telegram) notifyHandler(ctx tele.Context) error {
	// TODO: notify logic
	msg := "Настройка нотификации"
	return ctx.Edit(msg, showMeetings)
}

func (t *Telegram) cancelMeetingHandler(ctx tele.Context) error {
	// TODO: canceling meeting
	msg := "Отмена занятия"
	return ctx.Edit(msg, showMeetings)
}
