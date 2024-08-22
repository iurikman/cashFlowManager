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

	s.Run("PATCH", func() {
		s.Run("201/statusOK", func() {
			updatedWallet := new(models.Wallet)
			newWalletData := models.WalletDTO{
				Owner:    uuid.New(),
				Currency: "EURO",
				Balance:  100,
			}
			resp := s.sendRequest(
				context.Background(),
				http.MethodPatch,
				"/"+testWalletId.String(),
				newWalletData,
				&rest.HTTPResponse{Data: &updatedWallet},
			)
			s.Require().Equal(http.StatusOK, resp.StatusCode)
			s.Require().Equal(newWalletData.Owner, updatedWallet.Owner)
			s.Require().Equal(newWalletData.Currency, updatedWallet.Currency)
			s.Require().Equal(newWalletData.Balance, updatedWallet.Balance)
		})
		s.Run("400/statusBadRequest", func() {
			updatedWallet := new(models.Wallet)

			resp := s.sendRequest(
				context.Background(),
				http.MethodPatch,
				"/"+testWalletId.String(),
				"badRequest",
				&rest.HTTPResponse{Data: &updatedWallet},
			)
			s.Require().Equal(http.StatusBadRequest, resp.StatusCode)
		})
		s.Run("404/statusNotFound", func() {
			updatedWallet := new(models.Wallet)
			id := uuid.New().String()
			newWalletData := models.WalletDTO{
				Owner:    uuid.New(),
				Currency: "EURO",
				Balance:  100,
			}
			resp := s.sendRequest(
				context.Background(),
				http.MethodPatch,
				"/"+id,
				newWalletData,
				&rest.HTTPResponse{Data: &updatedWallet},
			)
			s.Require().Equal(http.StatusNotFound, resp.StatusCode)
		})
	})
	s.Run("DELETE", func() {
		s.Run("404/statusNotFound", func() {
			id := uuid.New()
			resp := s.sendRequest(
				context.Background(),
				http.MethodDelete,
				"/"+id.String(),
				nil,
				nil,
			)
			s.Require().Equal(http.StatusNotFound, resp.StatusCode)
		})
		s.Run("204/statusNoContent", func() {
			resp := s.sendRequest(
				context.Background(),
				http.MethodDelete,
				"/"+testWalletId.String(),
				nil,
				nil,
			)
			s.Require().Equal(http.StatusNoContent, resp.StatusCode)
		})
	})
}
