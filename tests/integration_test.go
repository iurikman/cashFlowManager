package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/iurikman/cashFlowManager/internal/config"
	"github.com/iurikman/cashFlowManager/internal/jwtgenerator"
	"github.com/iurikman/cashFlowManager/internal/rest"
	"github.com/iurikman/cashFlowManager/internal/service"
	"github.com/iurikman/cashFlowManager/internal/service/mocks"
	"github.com/iurikman/cashFlowManager/internal/store"
	_ "github.com/jackc/pgx/v5/stdlib"
	migrate "github.com/rubenv/sql-migrate"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

const bindAddress = "http://localhost:8080/api/v1/wallets"

type IntegrationTestSuite struct {
	suite.Suite
	cancel               context.CancelFunc
	store                *store.Postgres
	service              *service.Service
	server               *rest.Server
	authToken            string
	tokenGenerator       *jwtgenerator.JWTGenerator
	transactionsProducer *mocks.TransactionsProducer
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

func (s *IntegrationTestSuite) SetupSuite() {
	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel

	cfg := config.NewConfig()

	db, err := store.New(ctx, store.Config{
		PGUser:     cfg.PostgresUser,
		PGPassword: cfg.PostgresPassword,
		PGHost:     cfg.PostgresHost,
		PGPort:     cfg.PostgresPort,
		PGDatabase: cfg.PostgresDatabase,
	})
	s.Require().NoError(err)

	s.store = db

	err = s.store.Migrate(migrate.Up)
	s.Require().NoError(err)

	err = s.store.Truncate(ctx, "transactions_history", "wallets", "users")
	s.Require().NoError(err)

	xrConverter := MockConverter{}

	s.tokenGenerator = jwtgenerator.NewJWTGenerator()

	s.transactionsProducer = mocks.NewTransactionsProducer(s.T())
	s.transactionsProducer.On("ProduceTransaction", mock.Anything, mock.Anything).Return(nil)

	s.service = service.NewService(db, xrConverter, s.transactionsProducer)

	s.server, err = rest.NewServer(rest.ServerConfig{BindAddress: cfg.BindAddress}, s.service, s.tokenGenerator.GetPublicKey())
	s.Require().NoError(err)

	go func() {
		err = s.server.Start(ctx)
		s.Require().NoError(err)
	}()
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.cancel()
}

func (s *IntegrationTestSuite) sendRequest(ctx context.Context, method, endpoint string, body interface{}, dest interface{}) *http.Response {
	s.T().Helper()

	reqBody, err := json.Marshal(body)
	s.Require().NoError(err)

	req, err := http.NewRequestWithContext(ctx, method, bindAddress+endpoint, bytes.NewBuffer(reqBody))
	s.Require().NoError(err)

	req.Header.Set("Content-Type", "application/json")

	req.Header.Set("Authorization", "Bearer "+s.authToken)

	resp, err := http.DefaultClient.Do(req)
	s.Require().NoError(err)

	defer func() {
		err = resp.Body.Close()
		s.Require().NoError(err)
	}()

	if dest != nil {
		err = json.NewDecoder(resp.Body).Decode(&dest)
		s.Require().NoError(err)
	}

	return resp
}

func (s *IntegrationTestSuite) sendMetricsRequest(ctx context.Context, body interface{}, dest interface{}) *http.Response {
	s.T().Helper()

	reqBody, err := json.Marshal(body)
	s.Require().NoError(err)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost:8080/metrics", bytes.NewBuffer(reqBody))
	s.Require().NoError(err)

	req.Header.Set("Content-Type", "application/json")

	req.Header.Set("Authorization", "Bearer "+s.authToken)

	resp, err := http.DefaultClient.Do(req)
	s.Require().NoError(err)

	if resp.Body != nil {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			s.T().Errorf("Ошибка чтения тела ответа: %v", err)
		} else {
			fmt.Println(string(body)) // Вывод метрик в консоль
		}
	}

	defer func() {
		err = resp.Body.Close()
		s.Require().NoError(err)
	}()

	if dest != nil {
		err = json.NewDecoder(resp.Body).Decode(&dest)
		s.Require().NoError(err)
	}

	return resp
}

func toString(str string) *string {
	return &str
}
