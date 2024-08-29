package tests

import (
	"context"
	"math/rand"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/iurikman/cashFlowManager/internal/models"
	"github.com/iurikman/cashFlowManager/internal/rest"
)

const amountOfWallets = 20

func (s *IntegrationTestSuite) TestWallets() {
	s.ownersID = s.createListOfTestID(amountOfWallets)
	s.listOfWallets = s.createWallets(s.ownersID)

	s.Run("POST", func() {
		s.Run("201/statusCreated", func() {
			createdWallet, resp := s.testCreateWallet(s.ownersID[1])
			s.Require().Equal(http.StatusCreated, resp.StatusCode)
			s.Require().Equal(s.listOfWallets[1].Owner, createdWallet.Owner)
			s.Require().Equal(s.listOfWallets[1].Currency, createdWallet.Currency)
			s.Require().True(s.listOfWallets[1].Balance > 0)
		})

		s.Run("201/statusCreated(balance is zero)", func() {
			testWallet := models.Wallet{
				Owner:    s.listOfWallets[2].ID,
				Currency: "RUR",
				Balance:  0,
			}

			resp := s.sendRequest(
				context.Background(),
				http.MethodPost,
				"/",
				testWallet,
				&rest.HTTPResponse{Data: &testWallet},
			)
			s.Require().Equal(http.StatusCreated, resp.StatusCode)
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

		s.Run("400/StatusBadRequest(currency not allowed)", func() {
			testWallet := models.Wallet{
				Owner:    s.listOfWallets[2].ID,
				Currency: "NONECURRENCY",
				Balance:  10,
			}

			resp := s.sendRequest(
				context.Background(),
				http.MethodPost,
				"/",
				testWallet,
				nil,
			)
			s.Require().Equal(http.StatusBadRequest, resp.StatusCode)
		})

		s.Run("400/StatusBadRequest(balance below zero)", func() {
			testWallet := models.Wallet{
				Owner:    s.listOfWallets[2].ID,
				Currency: "",
				Balance:  -0.1,
			}

			resp := s.sendRequest(
				context.Background(),
				http.MethodPost,
				"/",
				testWallet,
				nil,
			)
			s.Require().Equal(http.StatusBadRequest, resp.StatusCode)
		})

		s.Run("400/StatusBadRequest(owner is empty)", func() {
			testWallet := models.Wallet{
				Owner:    uuid.Nil,
				Currency: "RUR",
				Balance:  10,
			}

			resp := s.sendRequest(
				context.Background(),
				http.MethodPost,
				"/",
				testWallet,
				nil,
			)
			s.Require().Equal(http.StatusBadRequest, resp.StatusCode)
		})
	})

	s.Run("GET", func() {
		s.Run("200/statusOK", func() {
			rWallet := new(models.Wallet)

			resp := s.sendRequest(
				context.Background(),
				http.MethodGet,
				"/"+s.listOfWallets[1].ID.String(),
				nil,
				&rest.HTTPResponse{Data: &rWallet},
			)
			s.Require().Equal(http.StatusOK, resp.StatusCode)
			s.Require().Equal(s.listOfWallets[1].ID, rWallet.ID)
			s.Require().Equal(s.listOfWallets[1].Owner, rWallet.Owner)
			s.Require().Equal(s.listOfWallets[1].Currency, rWallet.Currency)
			s.Require().Equal(s.listOfWallets[1].Balance, rWallet.Balance)
			s.Require().Equal(s.listOfWallets[1].Deleted, rWallet.Deleted)
		})

		s.Run("400/statusBadRequest", func() {
			resp := s.sendRequest(
				context.Background(),
				http.MethodGet,
				"/"+"badRequest",
				nil,
				nil,
			)
			s.Require().Equal(http.StatusBadRequest, resp.StatusCode)
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
		s.Run("200/statusOK", func() {
			updatedWallet := new(models.Wallet)
			newWalletData := models.WalletDTO{
				Owner:    s.ownersID[0],
				Currency: "EURO",
				Balance:  1000,
			}
			resp := s.sendRequest(
				context.Background(),
				http.MethodPatch,
				"/"+s.listOfWallets[1].ID.String(),
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
				"/"+s.listOfWallets[1].ID.String(),
				"badRequest",
				&rest.HTTPResponse{Data: &updatedWallet},
			)
			s.Require().Equal(http.StatusBadRequest, resp.StatusCode)
		})

		s.Run("404/statusNotFound", func() {
			updatedWallet := new(models.Wallet)
			id := uuid.New().String()
			newWalletData := models.WalletDTO{
				Owner:    s.ownersID[3],
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
		s.Run("204/statusNoContent", func() {
			resp := s.sendRequest(
				context.Background(),
				http.MethodDelete,
				"/"+s.listOfWallets[1].ID.String(),
				nil,
				nil,
			)
			s.Require().Equal(http.StatusNoContent, resp.StatusCode)
		})

		s.Run("400/statusBadRequest", func() {
			resp := s.sendRequest(
				context.Background(),
				http.MethodDelete,
				"/"+"badRequest",
				nil,
				nil,
			)
			s.Require().Equal(http.StatusBadRequest, resp.StatusCode)
		})

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
	})

	testWithdrawOperation := models.Transaction{
		TransactionID: uuid.New(),
		WalletID:      s.listOfWallets[2].ID,
		Amount:        100,
		Currency:      "CHY",
		OperationType: "ATM_withdraw",
	}

	testTransferOperation := models.Transaction{
		TransactionID:  uuid.New(),
		WalletID:       s.listOfWallets[2].ID,
		TargetWalletID: s.listOfWallets[3].ID,
		Amount:         50,
		Currency:       "CHY",
		OperationType:  "transfer",
	}

	testDepositOperation := models.Transaction{
		TransactionID: uuid.New(),
		WalletID:      s.listOfWallets[19].ID,
		Amount:        1000000,
		Currency:      "CHY",
		OperationType: "deposit",
	}

	testTargetUUIDNotFound := models.Transaction{
		TransactionID:  uuid.New(),
		WalletID:       s.listOfWallets[2].ID,
		TargetWalletID: uuid.New(),
		Amount:         50,
		Currency:       "CHY",
		OperationType:  "transfer",
	}

	testDeletedTrue := models.Transaction{
		TransactionID: uuid.New(),
		WalletID:      s.listOfWallets[1].ID,
		Amount:        400,
		Currency:      "AED",
		OperationType: "transfer",
	}

	testOwnerUUIDNotFound := models.Transaction{
		TransactionID: uuid.New(),
		WalletID:      uuid.New(),
		Amount:        400,
		Currency:      "AED",
		OperationType: "deposit",
	}

	testAmountBelowZero := models.Transaction{
		TransactionID: uuid.New(),
		WalletID:      s.listOfWallets[2].ID,
		Amount:        -1,
		Currency:      "AED",
		OperationType: "deposit",
	}

	testWalletIDIsNil := models.Transaction{
		TransactionID: uuid.New(),
		WalletID:      uuid.Nil,
		Amount:        10,
		Currency:      "AED",
		OperationType: "deposit",
	}

	testCurrencyNotAllowed := models.Transaction{
		TransactionID: uuid.New(),
		WalletID:      s.listOfWallets[0].ID,
		Amount:        10,
		Currency:      "NONECURRENCY",
		OperationType: "deposit",
	}

	s.Run("PUT", func() {
		s.Run("200/statusOK(deposit)", func() {
			executedTransaction := new(models.Transaction)
			resp := s.sendRequest(
				context.Background(),
				http.MethodPut,
				"/deposit",
				testDepositOperation,
				&rest.HTTPResponse{Data: &executedTransaction},
			)
			s.Require().Equal(http.StatusOK, resp.StatusCode)
		})

		s.Run("200/statusOK(transfer)", func() {
			executedTransaction := new(models.Transaction)
			resp := s.sendRequest(
				context.Background(),
				http.MethodPut,
				"/transfer",
				testTransferOperation,
				&rest.HTTPResponse{Data: &executedTransaction},
			)
			s.Require().Equal(http.StatusOK, resp.StatusCode)
		})

		s.Run("200/statusOK(withdraw)", func() {
			executedTransaction := new(models.Transaction)
			resp := s.sendRequest(
				context.Background(),
				http.MethodPut,
				"/withdraw",
				testWithdrawOperation,
				&rest.HTTPResponse{Data: &executedTransaction},
			)
			s.Require().Equal(http.StatusOK, resp.StatusCode)
		})

		s.Run("400/StatusBadRequest(deposit/amount below zero)", func() {
			executedTransaction := new(models.Transaction)
			resp := s.sendRequest(
				context.Background(),
				http.MethodPut,
				"/deposit",
				testAmountBelowZero,
				&rest.HTTPResponse{Data: &executedTransaction},
			)
			s.Require().Equal(http.StatusBadRequest, resp.StatusCode)
		})

		s.Run("400/StatusBadRequest(transfer/amount below zero)", func() {
			executedTransaction := new(models.Transaction)
			resp := s.sendRequest(
				context.Background(),
				http.MethodPut,
				"/transfer",
				testAmountBelowZero,
				&rest.HTTPResponse{Data: &executedTransaction},
			)
			s.Require().Equal(http.StatusBadRequest, resp.StatusCode)
		})

		s.Run("400/StatusBadRequest(withdraw/amount below zero)", func() {
			executedTransaction := new(models.Transaction)
			resp := s.sendRequest(
				context.Background(),
				http.MethodPut,
				"/withdraw",
				testAmountBelowZero,
				&rest.HTTPResponse{Data: &executedTransaction},
			)
			s.Require().Equal(http.StatusBadRequest, resp.StatusCode)
		})

		s.Run("400/StatusBadRequest(deposit/walletID is nil)", func() {
			executedTransaction := new(models.Transaction)
			resp := s.sendRequest(
				context.Background(),
				http.MethodPut,
				"/deposit",
				testWalletIDIsNil,
				&rest.HTTPResponse{Data: &executedTransaction},
			)
			s.Require().Equal(http.StatusBadRequest, resp.StatusCode)
		})

		s.Run("400/StatusBadRequest(transfer/walletID is nil)", func() {
			executedTransaction := new(models.Transaction)
			resp := s.sendRequest(
				context.Background(),
				http.MethodPut,
				"/transfer",
				testWalletIDIsNil,
				&rest.HTTPResponse{Data: &executedTransaction},
			)
			s.Require().Equal(http.StatusBadRequest, resp.StatusCode)
		})

		s.Run("400/StatusBadRequest(withdraw/walletID is nil)", func() {
			executedTransaction := new(models.Transaction)
			resp := s.sendRequest(
				context.Background(),
				http.MethodPut,
				"/withdraw",
				testWalletIDIsNil,
				&rest.HTTPResponse{Data: &executedTransaction},
			)
			s.Require().Equal(http.StatusBadRequest, resp.StatusCode)
		})

		s.Run("400/StatusBadRequest(deposit/currency not allowed)", func() {
			executedTransaction := new(models.Transaction)
			resp := s.sendRequest(
				context.Background(),
				http.MethodPut,
				"/deposit",
				testCurrencyNotAllowed,
				&rest.HTTPResponse{Data: &executedTransaction},
			)
			s.Require().Equal(http.StatusBadRequest, resp.StatusCode)
		})

		s.Run("400/StatusBadRequest(transfer/currency not allowed)", func() {
			executedTransaction := new(models.Transaction)
			resp := s.sendRequest(
				context.Background(),
				http.MethodPut,
				"/transfer",
				testCurrencyNotAllowed,
				&rest.HTTPResponse{Data: &executedTransaction},
			)
			s.Require().Equal(http.StatusBadRequest, resp.StatusCode)
		})

		s.Run("400/StatusBadRequest(withdraw/currency not allowed)", func() {
			executedTransaction := new(models.Transaction)
			resp := s.sendRequest(
				context.Background(),
				http.MethodPut,
				"/withdraw",
				testCurrencyNotAllowed,
				&rest.HTTPResponse{Data: &executedTransaction},
			)
			s.Require().Equal(http.StatusBadRequest, resp.StatusCode)
		})

		s.Run("400/StatusBadRequest(deposit/amount below zero)", func() {
			executedTransaction := new(models.Transaction)
			resp := s.sendRequest(
				context.Background(),
				http.MethodPut,
				"/deposit",
				testAmountBelowZero,
				&rest.HTTPResponse{Data: &executedTransaction},
			)
			s.Require().Equal(http.StatusBadRequest, resp.StatusCode)
		})

		s.Run("400/StatusBadRequest(transfer/amount below zero)", func() {
			executedTransaction := new(models.Transaction)
			resp := s.sendRequest(
				context.Background(),
				http.MethodPut,
				"/transfer",
				testAmountBelowZero,
				&rest.HTTPResponse{Data: &executedTransaction},
			)
			s.Require().Equal(http.StatusBadRequest, resp.StatusCode)
		})

		s.Run("400/StatusBadRequest(withdraw/amount below zero)", func() {
			executedTransaction := new(models.Transaction)
			resp := s.sendRequest(
				context.Background(),
				http.MethodPut,
				"/withdraw",
				testAmountBelowZero,
				&rest.HTTPResponse{Data: &executedTransaction},
			)
			s.Require().Equal(http.StatusBadRequest, resp.StatusCode)
		})

		s.Run("400/StatusBadRequest(operation type is empty)", func() {
			executedTransaction := new(models.Transaction)
			resp := s.sendRequest(
				context.Background(),
				http.MethodPut,
				"/withdraw",
				testAmountBelowZero,
				&rest.HTTPResponse{Data: &executedTransaction},
			)
			s.Require().Equal(http.StatusBadRequest, resp.StatusCode)
		})

		s.Run("404/StatusNotFound(ownerIDNotFound)", func() {
			executedTransaction := new(models.Transaction)
			resp := s.sendRequest(
				context.Background(),
				http.MethodPut,
				"/transfer",
				testOwnerUUIDNotFound,
				&rest.HTTPResponse{Data: &executedTransaction},
			)
			s.Require().Equal(http.StatusNotFound, resp.StatusCode)
		})

		s.Run("404/StatusNotFound(targetUUIDNotFound)", func() {
			executedTransaction := new(models.Transaction)
			resp := s.sendRequest(
				context.Background(),
				http.MethodPut,
				"/transfer",
				testTargetUUIDNotFound,
				&rest.HTTPResponse{Data: &executedTransaction},
			)
			s.Require().Equal(http.StatusNotFound, resp.StatusCode)
		})

		s.Run("404/StatusNotFound(deleted=true)", func() {
			executedTransaction := new(models.Transaction)
			resp := s.sendRequest(
				context.Background(),
				http.MethodPut,
				"/transfer",
				testDeletedTrue,
				&rest.HTTPResponse{Data: &executedTransaction},
			)
			s.Require().Equal(http.StatusNotFound, resp.StatusCode)
		})
	})
}

func (s *IntegrationTestSuite) testCreateWallet(testOwnerID uuid.UUID) (models.Wallet, http.Response) {
	createdWallet := new(models.Wallet)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	testWallet := models.Wallet{
		Owner:    testOwnerID,
		Currency: "RUR",
		Balance:  float64(r.Intn(1001)) / 10,
	}

	resp := s.sendRequest(
		context.Background(),
		http.MethodPost,
		"/",
		testWallet,
		&rest.HTTPResponse{Data: &createdWallet},
	)

	return *createdWallet, *resp
}

func (s *IntegrationTestSuite) createWallets(listOwnersID []uuid.UUID) []models.Wallet {
	var wallets []models.Wallet
	for _, id := range listOwnersID {
		newWallet, _ := s.testCreateWallet(id)
		wallets = append(wallets, newWallet)
	}

	return wallets
}

func (s *IntegrationTestSuite) createListOfTestID(amount int) []uuid.UUID {
	testWalletsID := make([]uuid.UUID, amount)
	for i := 0; i < amount; i++ {
		testWalletsID[i] = uuid.New()
	}

	return testWalletsID
}

func (s *IntegrationTestSuite) getRandomID(listOfTestWalletsID []uuid.UUID) uuid.UUID {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	return listOfTestWalletsID[r.Intn(len(listOfTestWalletsID))]
}
