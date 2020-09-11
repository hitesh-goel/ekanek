package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/cors"
)

var (
	errConfigInvalid = errors.New("invalid server config")
	errServe         = errors.New(`unable to serve`)
)

// Config represents the configuration necessary for this pkg.
type Config struct {
	CorsHeaders []string
	CorsMaxAge  int
	CorsMethods []string
	Port        int
	Timeout     time.Duration
}

func (c Config) isValid() bool {
	return c.Port != 0 && c.Timeout != time.Duration(0)
}

// Server wraps http.Server, which uses a sync.Mutex and sync.Once.
// As a result, when returning server, it's mandatory to return a pointer.
type Server struct {
	*cors.Cors
	http.Server
}

// HandleFunc ...
func (s *Server) HandleFunc(p string, h func(http.ResponseWriter, *http.Request)) {
	s.Server.Handler.(*http.ServeMux).HandleFunc(p, h)
}

// ListenAndServe enables CORS, runs a server,
// and attempts to shutdown gracefully,
// if certain signals are intercepted.
func (s *Server) ListenAndServe() error {
	s.enableCors()

	nc := make(chan os.Signal, 1)
	signal.Notify(nc, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)

	ec := make(chan error, 1)
	go func() {
		ec <- s.Server.ListenAndServe()
	}()

	select {
	case err := <-ec:
		return fmt.Errorf("%v: %w", errServe, err)
	case <-nc:
		ctx, cancel := context.WithTimeout(context.Background(), s.IdleTimeout)
		defer cancel()
		err := s.Shutdown(ctx)
		if err != nil {
			return s.Close()
		}
	}

	return nil
}

// New initializes a server.
func New(c Config) (*Server, error) {
	if !c.isValid() {
		return nil, errConfigInvalid
	}

	return &Server{
		cors.New(cors.Options{
			AllowedHeaders: c.CorsHeaders,
			AllowedMethods: c.CorsMethods,
		}),
		http.Server{
			Addr:         fmt.Sprintf(":%d", c.Port),
			Handler:      http.NewServeMux(),
			IdleTimeout:  c.Timeout,
			ReadTimeout:  c.Timeout,
			WriteTimeout: c.Timeout,
		},
	}, nil
}

func (s *Server) enableCors() {
	s.Server.Handler = s.Cors.Handler(s.Server.Handler)
}
