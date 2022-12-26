package main

import (
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pershin-daniil/TimeSlots/internal/rest"
	"github.com/pershin-daniil/TimeSlots/pkg/logger"
	"github.com/pershin-daniil/TimeSlots/pkg/store"
)

const (
	address = ":8080"
	version = "0.0.1"
)

var pgDSN = lookupEnv("PG_DSN", "postgres://postgres:secret@localhost:6431/timeslots?sslmode=disable")

func main() {
	log := logger.NewLogger()

	app, err := store.NewStore(log, pgDSN)
	if err != nil {
		log.Panic(err)
	}
	handler := rest.NewHandler(log, app, address, version)
	if err = handler.Run(); err != nil {
		log.Panic(err)
	}
}

func lookupEnv(key, defaultValue string) string {
	result := os.Getenv(key)
	if result == "" {
		return defaultValue
	}
	return result
}
