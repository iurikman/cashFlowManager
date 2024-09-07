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

const amountOfWallets = 4

func (s *IntegrationTestSuite) TestWallets() {
	testUserID := uuid.New()
	testUser := models.User{
		ID:       testUserID,
		Username: "testUser",
	}
	err := s.store.UpsertUser(context.Background(), testUser)
	s.Require().NoError(err)
	s.listOfWallets = s.createWallets(amountOfWallets, testUserID)

	s.Run("POST", func() {
		s.Run("201/statusCreated", func() {
			createdWallet, resp := s.testCreateWallet(testUserID)
			s.Require().Equal(http.StatusCreated, resp.StatusCode)
			s.Require().Equal(testUserID, createdWallet.Owner)
			s.Require().Equal(s.listOfWallets[1].Currency, "RUR")
			s.Require().True(createdWallet.Balance > 0)
		})

		s.Run("201/statusCreated(balance is zero)", func() {
			testWallet := models.Wallet{
				Owner:    testUserID,
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
				Owner:    testUserID,
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
				Owner:    testUserID,
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

		s.Run("400/StatusBadRequest(owner not found)", func() {
			id := uuid.New()
			testWallet := models.Wallet{
				Owner:    id,
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
			s.Require().Equal(testUserID, rWallet.Owner)
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
			id := uuid.New().String()
			resp := s.sendRequest(
				context.Background(),
				http.MethodGet,
				"/"+id,
				nil,
				nil,
			)
			s.Require().Equal(http.StatusNotFound, resp.StatusCode)
		})
	})

	s.Run("PATCH", func() {
		s.Run("200/statusOK", func() {
			updatedWallet := new(models.Wallet)
			newWalletData := models.Wallet{
				Owner:    testUserID,
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
			resp := s.sendRequest(
				context.Background(),
				http.MethodPatch,
				"/"+s.listOfWallets[1].ID.String(),
				"badRequest",
				nil,
			)
			s.Require().Equal(http.StatusBadRequest, resp.StatusCode)
		})

		s.Run("404/statusNotFound(wallet id not found)", func() {
			id := uuid.New().String()
			newWalletData := models.Wallet{
				Owner:    testUserID,
				Currency: "EURO",
				Balance:  100,
			}

			resp := s.sendRequest(
				context.Background(),
				http.MethodPatch,
				"/"+id,
				newWalletData,
				nil,
			)
			s.Require().Equal(http.StatusNotFound, resp.StatusCode)
		})

		s.Run("404/statusNotFound(owner id not found)", func() {
			newWalletData := models.Wallet{
				Owner:    uuid.New(),
				Currency: "EURO",
				Balance:  100,
			}

			resp := s.sendRequest(
				context.Background(),
				http.MethodPatch,
				"/"+"/"+s.listOfWallets[1].ID.String(),
				newWalletData,
				nil,
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
		Amount:        1,
		Currency:      "CHY",
		OperationType: "ATM_withdraw",
	}

	testTransferOperation := models.Transaction{
		TransactionID:  uuid.New(),
		WalletID:       s.listOfWallets[2].ID,
		TargetWalletID: s.listOfWallets[3].ID,
		Amount:         1,
		Currency:       "CHY",
		OperationType:  "transfer",
	}

	testTransferBalanceBelowZero := models.Transaction{
		TransactionID:  uuid.New(),
		WalletID:       s.listOfWallets[2].ID,
		TargetWalletID: s.listOfWallets[3].ID,
		Amount:         100000000,
		Currency:       "CHY",
		OperationType:  "transfer",
	}

	testWithdrawBalanceBelowZero := models.Transaction{
		TransactionID: uuid.New(),
		WalletID:      s.listOfWallets[3].ID,
		Amount:        100000000,
		Currency:      "CHY",
		OperationType: "withdraw",
	}

	testDepositOperation := models.Transaction{
		TransactionID: uuid.New(),
		WalletID:      s.listOfWallets[2].ID,
		Amount:        1000,
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

	idRUR := s.testCreateWalletForConverter(testUserID, "RUR", 10000.0)
	idAED := s.testCreateWalletForConverter(testUserID, "AED", 10000.0)

	testConverterWithdrawCHYfromRUR := models.Transaction{
		TransactionID: uuid.New(),
		WalletID:      idRUR,
		Amount:        10.0,
		Currency:      "CHY",
		OperationType: "withdraw CHY from RUR",
	}

	testConverterDepositCHYtoRUR := models.Transaction{
		TransactionID: uuid.New(),
		WalletID:      idRUR,
		Amount:        10.0,
		Currency:      "CHY",
		OperationType: "deposit CHY to RUR",
	}

	testConverterTransferAEDtoRUR := models.Transaction{
		TransactionID:  uuid.New(),
		WalletID:       idAED,
		TargetWalletID: idRUR,
		Amount:         10.0,
		Currency:       "AED",
		OperationType:  "transfer AED to RUR",
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

		s.Run("200/statusOK(test converter withdraw CHY from RUR)", func() {
			resp := s.sendRequest(
				context.Background(),
				http.MethodPut,
				"/withdraw",
				testConverterWithdrawCHYfromRUR,
				nil,
			)
			updatedWallet, _ := s.store.GetWalletByID(context.Background(), idRUR)
			s.Require().Equal(http.StatusOK, resp.StatusCode)
			s.Require().Equal(9990.0, updatedWallet.Balance)
		})

		s.Run("200/statusOK(test converter deposit CHY to RUR)", func() {
			resp := s.sendRequest(
				context.Background(),
				http.MethodPut,
				"/deposit",
				testConverterDepositCHYtoRUR,
				nil,
			)
			updatedWallet, _ := s.store.GetWalletByID(context.Background(), idRUR)
			s.Require().Equal(http.StatusOK, resp.StatusCode)
			s.Require().Equal(10000.0, updatedWallet.Balance)
		})

		s.Run("200/statusOK(test converter transfer AMD to RUR)", func() {
			resp := s.sendRequest(
				context.Background(),
				http.MethodPut,
				"/transfer",
				testConverterTransferAEDtoRUR,
				nil,
			)
			updatedWalletTo, _ := s.store.GetWalletByID(context.Background(), idRUR)
			updatedWalletFrom, _ := s.store.GetWalletByID(context.Background(), idAED)

			s.Require().Equal(http.StatusOK, resp.StatusCode)
			s.Require().Equal(10240.0, updatedWalletTo.Balance)
			s.Require().Equal(9990.0, updatedWalletFrom.Balance)
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
			resp := s.sendRequest(
				context.Background(),
				http.MethodPut,
				"/withdraw",
				testCurrencyNotAllowed,
				nil,
			)
			s.Require().Equal(http.StatusBadRequest, resp.StatusCode)
		})

		s.Run("400/StatusBadRequest(deposit/amount below zero)", func() {
			resp := s.sendRequest(
				context.Background(),
				http.MethodPut,
				"/deposit",
				testAmountBelowZero,
				nil,
			)
			s.Require().Equal(http.StatusBadRequest, resp.StatusCode)
		})

		s.Run("400/StatusBadRequest(transfer/amount below zero)", func() {
			resp := s.sendRequest(
				context.Background(),
				http.MethodPut,
				"/transfer",
				testAmountBelowZero,
				nil,
			)
			s.Require().Equal(http.StatusBadRequest, resp.StatusCode)
		})

		s.Run("400/StatusBadRequest(withdraw/amount below zero)", func() {
			resp := s.sendRequest(
				context.Background(),
				http.MethodPut,
				"/withdraw",
				testAmountBelowZero,
				nil,
			)
			s.Require().Equal(http.StatusBadRequest, resp.StatusCode)
		})

		s.Run("400/statusBadRequest(transfer/balance below zero)", func() {
			executedTransaction := new(models.Transaction)

			resp := s.sendRequest(
				context.Background(),
				http.MethodPut,
				"/transfer",
				testTransferBalanceBelowZero,
				&rest.HTTPResponse{Data: &executedTransaction},
			)
			s.Require().Equal(http.StatusBadRequest, resp.StatusCode)
		})

		s.Run("400/statusBadRequest(withdraw/balance below zero)", func() {
			executedTransaction := new(models.Transaction)

			resp := s.sendRequest(
				context.Background(),
				http.MethodPut,
				"/withdraw",
				testWithdrawBalanceBelowZero,
				&rest.HTTPResponse{Data: &executedTransaction},
			)
			s.Require().Equal(http.StatusBadRequest, resp.StatusCode)
		})

		s.Run("400/StatusBadRequest(operation type is empty)", func() {
			resp := s.sendRequest(
				context.Background(),
				http.MethodPut,
				"/withdraw",
				testAmountBelowZero,
				nil,
			)
			s.Require().Equal(http.StatusBadRequest, resp.StatusCode)
		})

		s.Run("404/StatusNotFound(ownerIDNotFound)", func() {
			resp := s.sendRequest(
				context.Background(),
				http.MethodPut,
				"/transfer",
				testOwnerUUIDNotFound,
				nil,
			)
			s.Require().Equal(http.StatusNotFound, resp.StatusCode)
		})

		s.Run("404/StatusNotFound(targetUUIDNotFound)", func() {
			resp := s.sendRequest(
				context.Background(),
				http.MethodPut,
				"/transfer",
				testTargetUUIDNotFound,
				nil,
			)
			s.Require().Equal(http.StatusNotFound, resp.StatusCode)
		})

		s.Run("404/StatusNotFound(deleted=true)", func() {
			resp := s.sendRequest(
				context.Background(),
				http.MethodPut,
				"/transfer",
				testDeletedTrue,
				nil,
			)
			s.Require().Equal(http.StatusNotFound, resp.StatusCode)
		})
	})
}

func (s *IntegrationTestSuite) testCreateWallet(testUserID uuid.UUID) (models.Wallet, http.Response) {
	createdWallet := new(models.Wallet)

	testWallet := models.Wallet{
		Owner:    testUserID,
		Currency: "RUR",
		Balance:  10.0,
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

func (s *IntegrationTestSuite) testCreateWalletForConverter(testOwnerID uuid.UUID, currency string, balance float64) uuid.UUID {
	var createdWallet models.Wallet
	testWallet := models.Wallet{
		Owner:    testOwnerID,
		Currency: currency,
		Balance:  balance,
	}

	s.sendRequest(
		context.Background(),
		http.MethodPost,
		"/",
		testWallet,
		&rest.HTTPResponse{Data: &createdWallet},
	)
	return createdWallet.ID
}

func (s *IntegrationTestSuite) createWallets(amountOfWallets int, testUserID uuid.UUID) []models.Wallet {
	var wallets []models.Wallet
	for i := 0; i < amountOfWallets; i++ {
		newWallet, _ := s.testCreateWallet(testUserID)
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
