package healthz

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

const (
	path = "/healthz"
)

var (
	errResponse = errors.New("unable to prepare response")
)

// Config represents the configuration necessary for this pkg.
type Config struct {
	ServiceName string
	StartupTime time.Time
}

type database struct {
	Available bool `json:"available"`
}

type response struct {
	Database    database `json:"database"`
	Message     string   `json:"message"`
	ServiceName string   `json:"service_name"`
	StartupTime string   `json:"startup_time"`
}

// Handle ...
func Handle(c Config, db *sql.DB) (string, func(http.ResponseWriter, *http.Request)) {
	return path, func(w http.ResponseWriter, r *http.Request) {
		err := handle(c, db, w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func handle(c Config, db *sql.DB, w http.ResponseWriter) error {
	err := db.Ping()
	if err != nil {
		return err
	}

	b, err := json.Marshal(response{
		Database: database{
			Available: true,
		},
		Message:     "OK",
		ServiceName: c.ServiceName,
		StartupTime: c.StartupTime.Format(time.RFC3339),
	})
	if err != nil {
		return fmt.Errorf("%v: %w", errResponse, err)
	}

	_, err = w.Write(b)
	if err != nil {
		return err
	}

	return nil
}
