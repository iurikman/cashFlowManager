package tests

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/iurikman/cashFlowManager/internal/models"
	"github.com/iurikman/cashFlowManager/internal/rest"
)

const amountOfWallets = 4

func (s *IntegrationTestSuite) TestWallets() {
	testUserID1 := uuid.New()
	testUser1 := models.User{
		ID:       testUserID1,
		Username: "testUser1",
		Email:    "testUser1@mail.com",
		Phone:    "1",
		Password: "password1",
	}
	authToken1, err := s.tokenGenerator.GetNewTokenString(testUser1)
	s.Require().NoError(err)
	err = s.store.UpsertUser(context.Background(), testUser1)
	s.Require().NoError(err)

	testUserID2 := uuid.New()
	testUser2 := models.User{
		ID:       testUserID2,
		Username: "testUser2",
		Email:    "testUser2@mail.com",
		Phone:    "2",
		Password: "password2",
	}
	authToken2, err := s.tokenGenerator.GetNewTokenString(testUser2)
	s.Require().NoError(err)
	err = s.store.UpsertUser(context.Background(), testUser2)
	s.Require().NoError(err)

	s.authToken = authToken2
	listOfWallets := s.createWallets(amountOfWallets, testUserID2)

	s.authToken = authToken1
	listOfWallets = append(s.createWallets(amountOfWallets, testUserID1), listOfWallets...)

	idRUR := s.createWalletForConverter(testUserID1, "RUR", 10000.0)
	idAED := s.createWalletForConverter(testUserID1, "AED", 10000.0)

	testAmountBelowZero := models.Transaction{
		TransactionID: uuid.New(),
		WalletID:      listOfWallets[2].ID,
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
		WalletID:      listOfWallets[0].ID,
		Amount:        10,
		Currency:      "NONECURRENCY",
		OperationType: "deposit",
	}

	s.Run("POST", func() {
		s.Run("201/statusCreated", func() {
			createdWallet, resp := s.testCreateWallet(testUserID1)
			s.Require().Equal(http.StatusCreated, resp.StatusCode)
			s.Require().Equal(testUserID1, createdWallet.Owner)
			s.Require().Equal(listOfWallets[0].Currency, "RUR")
			s.Require().Equal(createdWallet.Balance, 10.0)
		})

		s.Run("400/statusBadRequest", func() {
			resp := s.sendRequest(
				context.Background(),
				http.MethodPost,
				"/",
				"badRequest",
				nil,
			)
			s.Require().Equal(http.StatusBadRequest, resp.StatusCode)
		})

		s.Run("400/StatusBadRequest(currency not allowed)", func() {
			testWallet := models.Wallet{
				Owner:    testUserID1,
				Currency: "NONECURRENCY",
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

		s.Run("403/StatusForbidden(owner is empty)", func() {
			testWallet := models.Wallet{
				Owner:    uuid.Nil,
				Currency: "RUR",
			}

			resp := s.sendRequest(
				context.Background(),
				http.MethodPost,
				"/",
				testWallet,
				nil,
			)
			s.Require().Equal(http.StatusForbidden, resp.StatusCode)
		})

		s.Run("403/StatusForbidden(owner not found)", func() {
			id := uuid.New()
			testWallet := models.Wallet{
				Owner:    id,
				Currency: "RUR",
			}

			resp := s.sendRequest(
				context.Background(),
				http.MethodPost,
				"/",
				testWallet,
				nil,
			)
			s.Require().Equal(http.StatusForbidden, resp.StatusCode)
		})

		s.Run("401/StatusUnauthorized(empty token)", func() {
			s.authToken = ""
			resp := s.sendRequest(
				context.Background(),
				http.MethodPost,
				"/",
				nil,
				nil,
			)
			s.Require().Equal(http.StatusUnauthorized, resp.StatusCode)
			s.authToken = authToken1
		})

		s.Run("401/StatusUnauthorized(invalid token)", func() {
			s.authToken = "invalidToken"
			resp := s.sendRequest(
				context.Background(),
				http.MethodPost,
				"/",
				nil,
				nil,
			)
			s.Require().Equal(http.StatusUnauthorized, resp.StatusCode)
			s.authToken = authToken1
		})

		s.Run("403/StatusForbidden(token access closed)", func() {
			s.authToken = authToken2
			resp := s.sendRequest(
				context.Background(),
				http.MethodPost,
				"/",
				nil,
				nil,
			)
			s.Require().Equal(http.StatusForbidden, resp.StatusCode)
			s.authToken = authToken1
		})
	})

	s.Run("GET", func() {
		s.Run("200/statusOK", func() {
			rWallet := new(models.Wallet)

			resp := s.sendRequest(
				context.Background(),
				http.MethodGet,
				"/"+listOfWallets[0].ID.String(),
				nil,
				&rest.HTTPResponse{Data: &rWallet},
			)
			s.Require().Equal(http.StatusOK, resp.StatusCode)
			s.Require().Equal(listOfWallets[0].ID, rWallet.ID)
			s.Require().Equal(listOfWallets[0].Owner, rWallet.Owner)
			s.Require().Equal(listOfWallets[0].Currency, rWallet.Currency)
			s.Require().Equal(listOfWallets[0].Balance, rWallet.Balance)
			s.Require().Equal(listOfWallets[0].Deleted, rWallet.Deleted)
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
			rWallet := new(models.Wallet)
			walletDTO := models.WalletDTO{
				Name:     toString("new name"),
				Currency: toString("AED"),
			}
			resp := s.sendRequest(
				context.Background(),
				http.MethodPatch,
				"/"+listOfWallets[3].ID.String(),
				walletDTO,
				&rest.HTTPResponse{Data: &rWallet},
			)
			s.Require().Equal(http.StatusOK, resp.StatusCode)
			s.Require().Equal(listOfWallets[3].ID, rWallet.ID)
			s.Require().Equal(listOfWallets[3].Owner, rWallet.Owner)
			s.Require().Equal(*walletDTO.Name, rWallet.Name)
			s.Require().Equal(*walletDTO.Currency, rWallet.Currency)
			s.Require().Equal(listOfWallets[3].Deleted, rWallet.Deleted)
		})

		s.Run("200/StatusOK(change name only)", func() {
			rWallet := new(models.Wallet)
			walletDTO := models.WalletDTO{Name: toString("test name")}
			resp := s.sendRequest(
				context.Background(),
				http.MethodPatch,
				"/"+listOfWallets[0].ID.String(),
				walletDTO,
				&rest.HTTPResponse{Data: &rWallet},
			)
			s.Require().Equal(http.StatusOK, resp.StatusCode)
		})

		s.Run("200/StatusOK(change currency only)", func() {
			rWallet := new(models.Wallet)
			walletDTO := models.WalletDTO{Currency: toString("RUR")}
			resp := s.sendRequest(
				context.Background(),
				http.MethodPatch,
				"/"+listOfWallets[3].ID.String(),
				walletDTO,
				&rest.HTTPResponse{Data: &rWallet},
			)
			s.Require().Equal(http.StatusOK, resp.StatusCode)
		})

		s.Run("400/statusBadRequest(name is empty)", func() {
			walletDTO := models.WalletDTO{
				Name:     toString(""),
				Currency: toString("CHY"),
			}
			resp := s.sendRequest(
				context.Background(),
				http.MethodPatch,
				"/"+listOfWallets[3].ID.String(),
				walletDTO,
				nil,
			)
			s.Require().Equal(http.StatusBadRequest, resp.StatusCode)
		})

		s.Run("400/statusBadRequest(currency is empty)", func() {
			walletDTO := models.WalletDTO{
				Name:     toString("test name"),
				Currency: toString(""),
			}
			resp := s.sendRequest(
				context.Background(),
				http.MethodPatch,
				"/"+listOfWallets[3].ID.String(),
				walletDTO,
				nil,
			)
			s.Require().Equal(http.StatusBadRequest, resp.StatusCode)
		})

		s.Run("400/statusBadRequest(currency not allowed)", func() {
			rWallet := new(models.Wallet)
			walletDTO := models.WalletDTO{
				Name:     toString("test name"),
				Currency: toString("not allowed"),
			}
			resp := s.sendRequest(
				context.Background(),
				http.MethodPatch,
				"/"+listOfWallets[3].ID.String(),
				walletDTO,
				&rest.HTTPResponse{Data: &rWallet},
			)
			s.Require().Equal(http.StatusBadRequest, resp.StatusCode)
		})
	})

	s.Run("DELETE", func() {
		s.Run("204/StatusNoContent", func() {
			resp := s.sendRequest(
				context.Background(),
				http.MethodDelete,
				"/"+listOfWallets[3].ID.String(),
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

	testTransaction := new(models.Transaction)
	s.Run("PUT", func() {
		s.Run("200/statusOK(deposit)", func() {
			executedTransaction := new(models.Transaction)

			testDepositOperation := models.Transaction{
				TransactionID: uuid.New(),
				WalletID:      listOfWallets[0].ID,
				Amount:        1000,
				Currency:      "CHY",
				OperationType: "deposit",
			}

			resp := s.sendRequest(
				context.Background(),
				http.MethodPut,
				"/deposit",
				testDepositOperation,
				&rest.HTTPResponse{Data: &executedTransaction},
			)
			s.Require().Equal(http.StatusOK, resp.StatusCode)
			testTransaction = &testDepositOperation
		})

		s.Run("200/statusOK(transfer 1 CHY)", func() {
			executedTransaction := new(models.Transaction)

			testTransferOperation := models.Transaction{
				TransactionID:  uuid.New(),
				WalletID:       listOfWallets[0].ID,
				TargetWalletID: listOfWallets[1].ID,
				Amount:         1,
				Currency:       "CHY",
				OperationType:  "transfer",
			}

			resp := s.sendRequest(
				context.Background(),
				http.MethodPut,
				"/transfer",
				testTransferOperation,
				&rest.HTTPResponse{Data: &executedTransaction},
			)
			s.Require().Equal(http.StatusOK, resp.StatusCode)
		})

		s.Run("200/statusOK(withdraw 1 CHY)", func() {
			executedTransaction := new(models.Transaction)
			testWithdrawOperation := models.Transaction{
				TransactionID: uuid.New(),
				WalletID:      listOfWallets[0].ID,
				Amount:        1,
				Currency:      "CHY",
				OperationType: "withdraw",
			}
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
			testConverterWithdrawCHYfromRUR := models.Transaction{
				TransactionID: uuid.New(),
				WalletID:      idRUR,
				Amount:        10.0,
				Currency:      "CHY",
				OperationType: "withdraw",
			}

			resp := s.sendRequest(
				context.Background(),
				http.MethodPut,
				"/withdraw",
				testConverterWithdrawCHYfromRUR,
				nil,
			)
			updatedWallet, _ := s.store.GetWalletByID(context.Background(), idRUR, testUserID1)
			s.Require().Equal(http.StatusOK, resp.StatusCode)
			s.Require().Equal(9990.0, updatedWallet.Balance)
		})

		s.Run("200/statusOK(deposit/test converter deposit CHY to RUR)", func() {
			testTransaction := models.Transaction{
				TransactionID: uuid.New(),
				WalletID:      idRUR,
				Amount:        10.0,
				Currency:      "CHY",
				OperationType: "deposit",
			}

			resp := s.sendRequest(
				context.Background(),
				http.MethodPut,
				"/deposit",
				testTransaction,
				nil,
			)
			updatedWallet, _ := s.store.GetWalletByID(context.Background(), idRUR, testUserID1)

			s.Require().Equal(http.StatusOK, resp.StatusCode)
			s.Require().Equal(10000.0, updatedWallet.Balance)
		})

		s.Run("200/statusOK(transfer/test converter transfer AMD to RUR)", func() {
			testConverterTransferAEDtoRUR := models.Transaction{
				TransactionID:  uuid.New(),
				WalletID:       idAED,
				TargetWalletID: idRUR,
				Amount:         10.0,
				Currency:       "AED",
				OperationType:  "transfer",
			}

			resp := s.sendRequest(
				context.Background(),
				http.MethodPut,
				"/transfer",
				testConverterTransferAEDtoRUR,
				nil,
			)
			updatedWalletTo, _ := s.store.GetWalletByID(context.Background(), idRUR, testUserID1)
			updatedWalletFrom, _ := s.store.GetWalletByID(context.Background(), idAED, testUserID1)

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

		s.Run("400/StatusBadRequest(transfer/balance below zero)", func() {
			executedTransaction := new(models.Transaction)

			testTransferBalanceBelowZero := models.Transaction{
				TransactionID:  uuid.New(),
				WalletID:       listOfWallets[0].ID,
				TargetWalletID: listOfWallets[1].ID,
				Amount:         100000000,
				Currency:       "CHY",
				OperationType:  "transfer",
			}

			resp := s.sendRequest(
				context.Background(),
				http.MethodPut,
				"/transfer",
				testTransferBalanceBelowZero,
				&rest.HTTPResponse{Data: &executedTransaction},
			)
			s.Require().Equal(http.StatusBadRequest, resp.StatusCode)
		})

		s.Run("400/StatusBadRequest(withdraw/balance below zero)", func() {
			executedTransaction := new(models.Transaction)

			testWithdrawBalanceBelowZero := models.Transaction{
				TransactionID: uuid.New(),
				WalletID:      listOfWallets[0].ID,
				Amount:        100000000,
				Currency:      "CHY",
				OperationType: "withdraw",
			}
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
			testOwnerUUIDNotFound := models.Transaction{
				TransactionID: uuid.New(),
				WalletID:      uuid.New(),
				Amount:        400,
				Currency:      "AED",
				OperationType: "deposit",
			}

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
			testTargetUUIDNotFound := models.Transaction{
				TransactionID:  uuid.New(),
				WalletID:       listOfWallets[4].ID,
				TargetWalletID: uuid.New(),
				Amount:         50,
				Currency:       "CHY",
				OperationType:  "transfer",
			}

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
			testDeletedTrue := models.Transaction{
				TransactionID: uuid.New(),
				WalletID:      listOfWallets[0].ID,
				Amount:        400,
				Currency:      "AED",
				OperationType: "transfer",
			}

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

	s.Run("GET", func() {
		s.Run("/transactions", func() {
			s.Run("GET/200/statusOK", func() {
				transactions := new([]models.Transaction)
				params := "?limit=10&sorting=executed_at&descending=true"
				resp := s.sendRequest(
					context.Background(),
					http.MethodGet,
					"/"+listOfWallets[0].ID.String()+"/transactions"+params,
					nil,
					&rest.HTTPResponse{Data: &transactions},
				)
				s.Require().Equal(http.StatusOK, resp.StatusCode)

				for _, transaction := range *transactions {
					s.Require().Equal(listOfWallets[0].ID, transaction.WalletID)
				}

				s.Require().Equal(testTransaction.Amount, (*transactions)[2].Amount)
				s.Require().Equal(testTransaction.Currency, (*transactions)[2].Currency)
				s.Require().Equal(testTransaction.OperationType, (*transactions)[2].OperationType)
			})

			s.Run("422/StatusUnprocessableEntity", func() {
				params := "?limit=10&sorting=executed_at&descending=true"
				resp := s.sendRequest(
					context.Background(),
					http.MethodGet,
					"/badRequest"+"/transactions"+params,
					nil,
					nil,
				)
				s.Require().Equal(http.StatusUnprocessableEntity, resp.StatusCode)
			})

			s.Run("404/StatusNotFound", func() {
				params := "?limit=10&sorting=executed_at&descending=true"
				id := uuid.New()
				resp := s.sendRequest(
					context.Background(),
					http.MethodGet,
					"/"+id.String()+"/transactions"+params,
					nil,
					nil,
				)
				s.Require().Equal(http.StatusNotFound, resp.StatusCode)
			})

			s.Run("422/StatusUnprocessableEntity", func() {
				params := ""
				id := ""
				resp := s.sendRequest(
					context.Background(),
					http.MethodGet,
					"/"+id+"/transactions"+params,
					nil,
					nil,
				)
				s.Require().Equal(http.StatusUnprocessableEntity, resp.StatusCode)
			})
		})

		s.Run("/metrics", func() {
			resp := s.sendMetricsRequest(
				context.Background(),
				nil,
				nil,
			)
			s.Require().Equal(http.StatusOK, resp.StatusCode)
		})
	})
}

func (s *IntegrationTestSuite) testCreateWallet(testUserID uuid.UUID) (models.Wallet, http.Response) {
	createdWallet := new(models.Wallet)

	testWallet := models.Wallet{
		Owner:    testUserID,
		Name:     "Test Wallet",
		Currency: "RUR",
	}

	resp := s.sendRequest(
		context.Background(),
		http.MethodPost,
		"/",
		testWallet,
		&rest.HTTPResponse{Data: &createdWallet},
	)

	testTransaction := models.Transaction{
		TransactionID: uuid.New(),
		WalletID:      createdWallet.ID,
		OwnerID:       createdWallet.Owner,
		Amount:        10.0,
		Currency:      createdWallet.Currency,
		OperationType: "deposit",
	}

	s.sendRequest(
		context.Background(),
		http.MethodPut,
		"/deposit",
		testTransaction,
		nil,
	)

	wallet := new(models.Wallet)
	s.sendRequest(
		context.Background(),
		http.MethodGet,
		"/"+createdWallet.ID.String(),
		nil,
		&rest.HTTPResponse{Data: &wallet},
	)

	return *wallet, *resp
}

func (s *IntegrationTestSuite) createWalletForConverter(testOwnerID uuid.UUID, currency string, balance float64) uuid.UUID {
	var createdWallet models.Wallet
	testWallet := models.Wallet{
		Owner:    testOwnerID,
		Name:     "Test Wallet",
		Currency: currency,
	}

	s.sendRequest(
		context.Background(),
		http.MethodPost,
		"/",
		testWallet,
		&rest.HTTPResponse{Data: &createdWallet},
	)

	testTransaction := models.Transaction{
		TransactionID: uuid.New(),
		WalletID:      createdWallet.ID,
		OwnerID:       createdWallet.Owner,
		Amount:        balance,
		Currency:      currency,
		OperationType: "deposit",
	}

	s.sendRequest(
		context.Background(),
		http.MethodPut,
		"/deposit",
		testTransaction,
		nil,
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
