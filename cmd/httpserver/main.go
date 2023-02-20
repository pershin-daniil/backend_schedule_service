package main

import (
	"context"
	"github.com/pershin-daniil/TimeSlots/pkg/notifier"
	"github.com/pershin-daniil/TimeSlots/pkg/service"
	migrate "github.com/rubenv/sql-migrate"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pershin-daniil/TimeSlots/internal/rest"
	"github.com/pershin-daniil/TimeSlots/pkg/logger"
	"github.com/pershin-daniil/TimeSlots/pkg/pgstore"
)

const (
	address = ":8080"
	version = "0.0.1"
)

var pgDSN = lookupEnv("PG_DSN", "postgres://postgres:secret@localhost:6431/timeslots?sslmode=disable")

func main() {
	log := logger.NewLogger()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	store, err := pgstore.NewStore(ctx, log, pgDSN)
	if err != nil {
		log.Panic(err)
	}
	if err = store.Migrate(migrate.Up); err != nil {
		log.Panic(err)
	}
	dummy := notifier.NewDummyNotifier(log)
	app := service.NewScheduleService(log, store, dummy)
	server := rest.NewServer(log, app, address, version)
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGHUP)
		<-sigCh
		log.Info("Received signal, shutting down...")
		cancel()
	}()
	if err = server.Run(ctx); err != nil {
		log.Panic(err)
	}
	log.Info("Server stopped")
}

func lookupEnv(key, defaultValue string) string {
	result := os.Getenv(key)
	if result == "" {
		return defaultValue
	}
	return result
}
