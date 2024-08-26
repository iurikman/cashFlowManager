package tests

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/iurikman/cashFlowManager/internal/models"
	"github.com/iurikman/cashFlowManager/internal/rest"
)

var (
	testWalletId = uuid.New()
	testOwnerId  = uuid.New()
)

func (s *IntegrationTestSuite) TestWallets() {
	testWallet1 := models.Wallet{
		ID:        testWalletId,
		Owner:     testOwnerId,
		Currency:  "RUR",
		Balance:   198,
		CreatedAt: time.Now(),
		Deleted:   false,
	}
	s.Run("POST", func() {
		s.Run("201/statusCreated", func() {
			rWallet := new(models.Wallet)

			resp := s.sendRequest(
				context.Background(),
				http.MethodPost,
				"/",
				testWallet1,
				&rest.HTTPResponse{Data: &rWallet},
			)

			s.Require().Equal(http.StatusCreated, resp.StatusCode)
			s.Require().Equal(testWalletId, rWallet.ID)
			s.Require().Equal(testWallet1.Owner, rWallet.Owner)
			s.Require().Equal(testWallet1.Currency, rWallet.Currency)
			s.Require().Equal(testWallet1.Balance, rWallet.Balance)
			s.Require().Equal(testWallet1.Deleted, rWallet.Deleted)
		})

		s.Run("409/statusConflict", func() {
			rWallet := new(models.Wallet)

			resp := s.sendRequest(
				context.Background(),
				http.MethodPost,
				"/",
				testWallet1,
				&rest.HTTPResponse{Data: &rWallet},
			)
			s.Require().Equal(http.StatusConflict, resp.StatusCode)
		})

		s.Run("400/statusBadRequest", func() {
			rWallet := new(models.Wallet)
			resp := s.sendRequest(
				context.Background(),
				http.MethodPost,
				"/",
				"badRequest",
				&rest.HTTPResponse{Data: &rWallet},
			)
			s.Require().Equal(http.StatusBadRequest, resp.StatusCode)
		})
	})

	s.Run("GET", func() {
		s.Run("201/statusOK", func() {
			rWallet := new(models.Wallet)
			id := testWalletId.String()
			resp := s.sendRequest(
				context.Background(),
				http.MethodGet,
				"/"+id,
				nil,
				&rest.HTTPResponse{Data: &rWallet},
			)
			s.Require().Equal(http.StatusOK, resp.StatusCode)
			s.Require().Equal(testWalletId, rWallet.ID)
			s.Require().Equal(testWallet1.Owner, rWallet.Owner)
			s.Require().Equal(testWallet1.Currency, rWallet.Currency)
			s.Require().Equal(testWallet1.Balance, rWallet.Balance)
			s.Require().Equal(testWallet1.Deleted, rWallet.Deleted)
		})
		s.Run("404/statusNotFound", func() {
			rWallet := new(models.Wallet)
			id := uuid.New().String()
			resp := s.sendRequest(
				context.Background(),
				http.MethodGet,
				"/"+id,
				nil,
				&rest.HTTPResponse{Data: &rWallet},
			)
			s.Require().Equal(http.StatusNotFound, resp.StatusCode)
		})
	})
}
