package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/hitesh-goel/ekanek/internal/handlers/assets"
	"github.com/hitesh-goel/ekanek/internal/handlers/auth"
	"log"
	"time"

	"github.com/hitesh-goel/ekanek/internal/db"
	"github.com/hitesh-goel/ekanek/internal/handlers/healthz"
	"github.com/hitesh-goel/ekanek/internal/handlers/user"
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
		AWSRegion:  flag.String("aws-region", "", "AWS Region"),
		AWSKey:     flag.String("aws-key", "", "AWS Key"),
		AWSSecret:  flag.String("aws-secret", "", "AWS Secret"),
		PrivateKey: flag.String("private-key", "", "Secreet Key"),
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
	AWSRegion  *string
	AWSKey     *string
	AWSSecret  *string
	PrivateKey *string
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

	// TODO: Handle Endpoint & S3ForcePathStyle for local development using environment variable
	ar := assets.AssetResources{
		Session: session.Must(session.NewSession(&aws.Config{
			Credentials:      credentials.NewStaticCredentials(*cfg.AWSKey, *cfg.AWSSecret, ""),
			S3ForcePathStyle: aws.Bool(true),
			Region:           aws.String(*cfg.AWSRegion),
			Endpoint:         aws.String("http://s3-fake:4572"),
		})),
		DTO: db,
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

	srv.HandleFunc(user.HandleSignup(*cfg.PrivateKey, db))
	srv.HandleFunc(user.HandleLogin(*cfg.PrivateKey, db))
	srv.HandleFunc(auth.Auth(assets.HandleAssetUpload(&ar)))
	srv.HandleFunc(auth.Auth(assets.HandleListAssets(&ar)))
	srv.HandleFunc(auth.Auth(assets.HandlePublicAsset(&ar)))
	srv.HandleFunc(auth.Auth(assets.HandleDeleteAsset(&ar)))
	srv.HandleFunc(assets.HandleAssetDownload(&ar))

	logger.Info().Msg("listening...")
	return srv.ListenAndServe()
}
