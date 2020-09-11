package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/hitesh-goel/ekanek/internal/db"
	"github.com/hitesh-goel/ekanek/internal/handlers/healthz"
	"github.com/hitesh-goel/ekanek/internal/logging"
	"github.com/hitesh-goel/ekanek/internal/server"
)

var (
	cfg = config{
		DbHost:     flag.String("db-host", "", "DB host"),
		DbName:     flag.String("db-name", "", "DB name"),
		DbPass:     flag.String("db-pass", "", "DB password"),
		DbPort:     flag.Int("db-port", 0, "DB port"),
		DbUser:     flag.String("db-user", "", "DB user"),
		LogLevel:   flag.String("log-level", "", "Logger level"),
		SrvTimeout: flag.Duration("srv-timeout", time.Duration(0), "Server timeout (e.g., 10s)"),
	}

	errRun = errors.New("unable to run")
)

type config struct {
	DbHost     *string
	DbName     *string
	DbPass     *string
	DbPort     *int
	DbUser     *string
	LogLevel   *string
	SrvTimeout *time.Duration
}

func init() {
	flag.Parse()
}

func main() {
	err := run()
	if err != nil {
		log.Fatal(err)
	}
}

func run() error {
	logger := logging.New(logging.Config{
		Level: *cfg.LogLevel,
	})

	db, err := db.New(db.Config{
		Host:     *cfg.DbHost,
		Name:     *cfg.DbName,
		Password: *cfg.DbPass,
		User:     *cfg.DbUser,
	})
	if err != nil {
		return fmt.Errorf("%v: %w", errRun, err)
	}

	if err != nil {
		return fmt.Errorf("%v: %w", errRun, err)
	}

	srv, err := server.New(server.Config{
		CorsHeaders: []string{"Accept,Content-Length", "Content-Type", "Authorization"},
		CorsMethods: []string{"GET,", "POST", "PUT", "OPTIONS", "DELETE"},
		Port:        8080,
		Timeout:     *cfg.SrvTimeout,
	})
	if err != nil {
		return fmt.Errorf("%v: %w", errRun, err)
	}

	srv.HandleFunc(healthz.Handle(healthz.Config{
		ServiceName: "ekanek",
		StartupTime: time.Now().UTC(),
	}, db))

	logger.Info().Msg("listening...")
	return srv.ListenAndServe()
}
