package logger

import "github.com/sirupsen/logrus"

func New() *logrus.Logger {
	log := logrus.New()
	log.SetLevel(logrus.DebugLevel)
	return log
}
