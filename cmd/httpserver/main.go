package main

import (
	"github.com/pershin-daniil/TimeSlots/internal/rest"
	"github.com/pershin-daniil/TimeSlots/pkg/logger"
)

const (
	address = ":8080"
	version = "0.0.1"
)

func main() {
	log := logger.NewLogger()
	handler := rest.NewHandler(log, address, version)
	if err := handler.Run(); err != nil {
		log.Panic(err)
	}
}
