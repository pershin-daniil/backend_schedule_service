package telegram

import tele "gopkg.in/telebot.v3"

func (t *Telegram) initButtons() {
	registration.Inline(
		registration.Row(registrationBtn))
	availMeetings.Inline(
		availMeetings.Row(availableMeetingsBtn),
		showMeetings.Row(myMeetingBtn))
	showMeetings.Inline(
		showMeetings.Row(myMeetingBtn))
	settings.Inline(
		availMeetings.Row(availableMeetingsBtn),
		settings.Row(notificationBtn),
		settings.Row(cancelMeetingBtn))
}

var (
	registration    = &tele.ReplyMarkup{}
	registrationBtn = registration.Data("Продолжить", "continue")
)

var (
	availMeetings        = &tele.ReplyMarkup{}
	availableMeetingsBtn = availMeetings.Data("Записаться на тренировку", "schedule")
)

var (
	showMeetings = &tele.ReplyMarkup{}
	myMeetingBtn = showMeetings.Data("Мои тренировки", "trainings")
)

var (
	settings         = &tele.ReplyMarkup{}
	notificationBtn  = settings.Data("Напоминание", "notify")
	cancelMeetingBtn = settings.Data("Отмена тренировки", "cancel")
)
