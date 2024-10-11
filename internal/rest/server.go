package rest

import (
	"context"
	"crypto/rsa"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

type ServerConfig struct {
	BindAddress string
}

const (
	readHeaderTimeout       = 10 * time.Second
	maxHeaderBytes          = 1 << 20
	gracefulShutdownTimeout = 5 * time.Second
)

type Server struct {
	serverConfig ServerConfig
	service      service
	key          *rsa.PublicKey
	router       *chi.Mux
	server       *http.Server
	metrics      *metrics
}

func NewServer(serverConfig ServerConfig, srv service, key *rsa.PublicKey) (*Server, error) {
	router := chi.NewRouter()

	return &Server{
		serverConfig: serverConfig,
		service:      srv,
		router:       router,
		key:          key,
		server: &http.Server{
			Addr:              serverConfig.BindAddress,
			Handler:           router,
			ReadHeaderTimeout: readHeaderTimeout,
			MaxHeaderBytes:    maxHeaderBytes,
		},
		metrics: newMetrics(),
	}, nil
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
	s.router.Route("/api", func(r chi.Router) {
		r.Use(s.jwtAuth)

		r.Route("/v1", func(r chi.Router) {
			r.Route("/wallets", func(r chi.Router) {
				r.Post("/", s.createWallet)
				r.Get("/{id}", s.getWalletByID)
				r.Patch("/{id}", s.updateWallet)
				r.Delete("/{id}", s.deleteWallet)

				r.Put("/withdraw", s.withdraw)
				r.Put("/transfer", s.transfer)
				r.Put("/deposit", s.deposit)

				r.Get("/{id}/transactions", s.getTransactions)
			})
		})
	})

	s.router.Get("/metrics", promhttp.Handler().ServeHTTP)
}
