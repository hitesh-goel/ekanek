// Package db provides support for access the database.
package db

import (
	"database/sql"
	"errors"
	"fmt"
	"net/url"

	_ "github.com/lib/pq" // The database driver in use.
)

const (
	driverName = "postgres"
)

var (
	errConfigInvalid = errors.New("invalid db config")
)

// Config is the required properties to use the database.
type Config struct {
	User     string
	Password string
	Host     string
	Name     string
}

func (c Config) isValid() bool {
	return c.Host != "" && c.Name != "" && c.Password != "" && c.User != ""
}

// New initializes a database abstraction.
func New(c Config) (*sql.DB, error) {
	if !c.isValid() {
		fmt.Println(c)
		return nil, errConfigInvalid
	}

	sslMode := "disable"

	q := make(url.Values)
	q.Set("sslmode", sslMode)
	q.Set("timezone", "utc")

	u := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(c.User, c.Password),
		Host:     c.Host,
		Path:     c.Name,
		RawQuery: q.Encode(),
	}

	db, err := sql.Open(driverName, u.String())
	if err != nil {
		return nil, err
	}

	return db, nil
}
