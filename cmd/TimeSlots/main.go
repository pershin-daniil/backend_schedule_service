package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/pershin-daniil/TimeSlots/internal/calendar"

	"github.com/pershin-daniil/TimeSlots/pkg/notifier"

	"github.com/pershin-daniil/TimeSlots/pkg/worker"

	"github.com/pershin-daniil/TimeSlots/internal/telegram"

	"github.com/pershin-daniil/TimeSlots/pkg/service"
	migrate "github.com/rubenv/sql-migrate"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pershin-daniil/TimeSlots/internal/rest"
	"github.com/pershin-daniil/TimeSlots/pkg/logger"
	"github.com/pershin-daniil/TimeSlots/pkg/pgstore"
)

const (
	address = ":8080"
	version = "0.0.1"
)

var (
	pgDSN   = lookupEnv("PG_DSN", "postgres://postgres:secret@localhost:6431/timeslots?sslmode=disable")
	tgToken = os.Getenv("TG_TOKEN")
)

func main() {
	log := logger.New()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	store, err := pgstore.New(ctx, log, pgDSN)
	if err != nil {
		log.Panic(err)
	}
	if err = store.Migrate(migrate.Up); err != nil {
		log.Panic(err)
	}
	tgBot, err := telegram.NewBot(tgToken)
	if err != nil {
		log.Panic(err)
	}
	ntf := notifier.New(log, tgBot)
	cal := calendar.New(ctx, log)
	app := service.NewScheduleService(log, store)
	tg, err := telegram.New(log, tgBot, app, cal)
	if err != nil {
		log.Panic(err)
	}
	server := rest.New(log, app, address, version)
	notifyUsers := worker.New(log, store, ntf)

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGHUP)
		<-sigCh
		log.Info("Received signal, shutting down...")
		cancel()
	}()
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		tg.Run(ctx)
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err = server.Run(ctx); err != nil {
			log.Panic(err)
		}
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err = notifyUsers.SendNotificationBeforeTraining(ctx); err != nil {
			log.Panic(err)
		}
	}()
	wg.Wait()
}

func lookupEnv(key, defaultValue string) string {
	result := os.Getenv(key)
	if result == "" {
		return defaultValue
	}
	return result
}
