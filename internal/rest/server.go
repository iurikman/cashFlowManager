package rest

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/sirupsen/logrus"
)

const (
	readHeaderTimeout       = 10 * time.Second
	maxHeaderBytes          = 1 << 20
	gracefulShutdownTimeout = 5 * time.Second
)

type Config struct {
	BindAddress string
}

type Server struct {
	router  *chi.Mux
	service service
	server  *http.Server
}

func NewServer(config Config, service service) *Server {
	router := chi.NewRouter()

	return &Server{
		service: service,
		server: &http.Server{
			Addr:              config.BindAddress,
			Handler:           router,
			ReadHeaderTimeout: readHeaderTimeout,
			MaxHeaderBytes:    maxHeaderBytes,
		},
	}
}

func (s *Server) Start(ctx context.Context) error {
	s.configRouter()

	go func() {
		<-ctx.Done()
		ctxWithTimeout, cancel := context.WithTimeout(ctx, gracefulShutdownTimeout)

		defer cancel()

		err := s.server.Shutdown(ctxWithTimeout)
		if err != nil {
			logrus.Warnf("failed to shutdown gracefully %s", err)
		}
	}()

	err := s.server.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("s.server.ListenAndServe() err: %w", err)
	}

	return nil
}

func (s *Server) configRouter() {
	//	s.router.Route("/api/v1", func(r chi.Router) {
	//		r.Post("/", s.createWallet)
	//	})
}
