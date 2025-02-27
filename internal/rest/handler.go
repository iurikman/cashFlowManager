package rest

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/gorilla/schema"
	"github.com/iurikman/cashFlowManager/internal/models"
	log "github.com/sirupsen/logrus"
)

const standartPage = 10

type service interface {
	CreateWallet(context context.Context, wallet models.Wallet) (*models.Wallet, error)
	GetWalletByID(ctx context.Context, id, ownerID uuid.UUID) (*models.Wallet, error)
	UpdateWallet(ctx context.Context, walletID, ownerID uuid.UUID, walletDTO models.WalletDTO) (*models.Wallet, error)
	DeleteWallet(context context.Context, id, ownerID uuid.UUID) error
	Deposit(ctx context.Context, transaction models.Transaction, ownerID uuid.UUID) error
	Transfer(ctx context.Context, transaction models.Transaction, ownerID uuid.UUID) error
	Withdraw(ctx context.Context, transaction models.Transaction, ownerID uuid.UUID) error
	GetTransactions(ctx context.Context, ID uuid.UUID, params models.Params) ([]*models.Transaction, error)
}

type HTTPResponse struct {
	Data  any    `json:"data"`
	Error string `json:"error"`
}

func (s *Server) createWallet(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	defer func() {
		s.metrics.requestsDuration.WithLabelValues("createWallet", r.URL.Path).Observe(time.Since(startTime).Seconds())
	}()

	var wallet models.Wallet

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewDecoder(r.Body).Decode(&wallet); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, err.Error())

		return
	}

	ownerID := s.getOwnerIDFromRequest(r)

	if ownerID != wallet.Owner {
		writeErrorResponse(w, http.StatusForbidden, "forbidden")

		return
	}

	if err := wallet.Validate(); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, err.Error())

		return
	}

	createdUser, err := s.service.CreateWallet(r.Context(), wallet)

	switch {
	case errors.Is(err, models.ErrUserNotFound):
		writeErrorResponse(w, http.StatusBadRequest, err.Error())

		return
	case err != nil:
		writeErrorResponse(w, http.StatusInternalServerError, "internal server error")
		log.Warnf("failed to create new wallet: %v", err)

		return
	}

	writeOkResponse(w, http.StatusCreated, createdUser)
}

func (s *Server) getWalletByID(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	defer func() {
		s.metrics.requestsDuration.WithLabelValues("getWalletByID", r.URL.Path).Observe(time.Since(startTime).Seconds())
	}()

	walletID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeErrorResponse(w, http.StatusBadRequest, err.Error())

		return
	}

	ownerID := s.getOwnerIDFromRequest(r)

	wallet, err := s.service.GetWalletByID(r.Context(), walletID, ownerID)

	switch {
	case errors.Is(err, models.ErrWalletNotFound):
		writeErrorResponse(w, http.StatusNotFound, err.Error())

		return
	case err != nil:
		writeErrorResponse(w, http.StatusInternalServerError, "internal server error")

		return
	}

	writeOkResponse(w, http.StatusOK, wallet)
}

func (s *Server) updateWallet(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	defer func() {
		s.metrics.requestsDuration.WithLabelValues("updateWallet", r.URL.Path).Observe(time.Since(startTime).Seconds())
	}()

	var walletDTO models.WalletDTO

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewDecoder(r.Body).Decode(&walletDTO); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, err.Error())

		return
	}

	if err := walletDTO.Validate(); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, err.Error())

		return
	}

	walletID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeErrorResponse(w, http.StatusBadRequest, err.Error())

		return
	}

	ownerID := s.getOwnerIDFromRequest(r)

	wallet, err := s.service.UpdateWallet(r.Context(), walletID, ownerID, walletDTO)
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "internal server error")
		log.Warnf("failed to update wallet: %v", err)

		return
	}

	writeOkResponse(w, http.StatusOK, wallet)
}

func (s *Server) deleteWallet(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	defer func() {
		s.metrics.requestsDuration.WithLabelValues("deleteWallet", r.URL.Path).Observe(time.Since(startTime).Seconds())
	}()

	walletID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeErrorResponse(w, http.StatusBadRequest, err.Error())

		return
	}

	ownerID := s.getOwnerIDFromRequest(r)

	err = s.service.DeleteWallet(r.Context(), walletID, ownerID)

	switch {
	case errors.Is(err, models.ErrWalletNotFound):
		writeErrorResponse(w, http.StatusNotFound, err.Error())

		return
	case err != nil:
		writeErrorResponse(w, http.StatusInternalServerError, "internal server error")
		log.Warnf("failed to delete wallet: %v", err)

		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) deposit(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	defer func() {
		s.metrics.requestsDuration.WithLabelValues("deposit", r.URL.Path).Observe(time.Since(startTime).Seconds())
	}()

	var transaction models.Transaction

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewDecoder(r.Body).Decode(&transaction); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, err.Error())

		return
	}

	if err := transaction.Validate(); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, err.Error())

		return
	}

	ownerID := s.getOwnerIDFromRequest(r)

	err := s.service.Deposit(r.Context(), transaction, ownerID)

	switch {
	case errors.Is(err, models.ErrWalletNotFound):
		writeErrorResponse(w, http.StatusNotFound, err.Error())

		return
	case err != nil:
		writeErrorResponse(w, http.StatusInternalServerError, "internal server error")
		log.Warnf("failed to deposit transaction: %v", err)

		return
	}

	writeOkResponse(w, http.StatusOK, nil)
}

//nolint:dupl
func (s *Server) transfer(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	defer func() {
		s.metrics.requestsDuration.WithLabelValues("transfer", r.URL.Path).Observe(time.Since(startTime).Seconds())
	}()

	var transaction models.Transaction

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewDecoder(r.Body).Decode(&transaction); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, err.Error())

		return
	}

	ownerID := s.getOwnerIDFromRequest(r)

	if err := transaction.Validate(); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, err.Error())

		return
	}

	err := s.service.Transfer(r.Context(), transaction, ownerID)

	switch {
	case errors.Is(err, models.ErrWalletNotFound):
		writeErrorResponse(w, http.StatusNotFound, err.Error())

		return
	case errors.Is(err, models.ErrBalanceBelowZero):
		writeErrorResponse(w, http.StatusBadRequest, err.Error())

		return
	case err != nil:
		writeErrorResponse(w, http.StatusInternalServerError, "internal server error")
		log.Warnf("failed to transfer transaction: %v", err)

		return
	}

	writeOkResponse(w, http.StatusOK, nil)
}

//nolint:dupl
func (s *Server) withdraw(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	defer func() {
		s.metrics.requestsDuration.WithLabelValues("withdraw", r.URL.Path).Observe(time.Since(startTime).Seconds())
	}()

	var transaction models.Transaction

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewDecoder(r.Body).Decode(&transaction); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, err.Error())

		return
	}

	ownerID := s.getOwnerIDFromRequest(r)

	if err := transaction.Validate(); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, err.Error())

		return
	}

	err := s.service.Withdraw(r.Context(), transaction, ownerID)

	switch {
	case errors.Is(err, models.ErrWalletNotFound):
		writeErrorResponse(w, http.StatusNotFound, err.Error())

		return
	case errors.Is(err, models.ErrBalanceBelowZero):
		writeErrorResponse(w, http.StatusBadRequest, err.Error())

		return
	case err != nil:
		writeErrorResponse(w, http.StatusInternalServerError, "internal server error")
		log.Warnf("failed to withdraw transaction: %v", err)

		return
	}

	writeOkResponse(w, http.StatusOK, nil)
}

func (s *Server) getTransactions(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	defer func() {
		s.metrics.requestsDuration.WithLabelValues("getTransactions", r.URL.Path).Observe(time.Since(startTime).Seconds())
	}()

	params, err := parseParams(r.URL.Query())
	if err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "invalid query parameters")

		return
	}

	sID := chi.URLParam(r, "id")

	if sID == " " {
		writeErrorResponse(w, http.StatusUnprocessableEntity, "id is empty")

		return
	}

	id, err := uuid.Parse(sID)
	if err != nil {
		writeErrorResponse(w, http.StatusUnprocessableEntity, "invalid wallet id")

		return
	}

	transactions, err := s.service.GetTransactions(r.Context(), id, *params)

	switch {
	case errors.Is(err, models.ErrTransactionsNotFound):
		writeErrorResponse(w, http.StatusNotFound, "wallet not found")

		return
	case err != nil:
		writeErrorResponse(w, http.StatusInternalServerError, "internal server error")
		log.Warnf("failed to get transactions: %v", err)

		return
	}

	writeOkResponse(w, http.StatusOK, transactions)
}

func parseParams(query url.Values) (*models.Params, error) {
	var params models.Params

	if err := schema.NewDecoder().Decode(&params, query); err != nil {
		return nil, fmt.Errorf("schema.NewDecoder().Decode(params, query) err: %w", err)
	}

	if params.Limit == 0 {
		params.Limit = standartPage
	}

	return &params, nil
}

func writeOkResponse(w http.ResponseWriter, statusCode int, respData any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(HTTPResponse{Data: respData}); err != nil {
		log.Warnf("json.NewEncoder(w).Encode(HTTPResponse{Data: respData}) err: %v", err)
	}
}

func writeErrorResponse(w http.ResponseWriter, statusCode int, description string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(HTTPResponse{Error: description}); err != nil {
		log.Warnf("json.NewEncoder(w).Encode(HTTPResponse{Error: description}) err: %s", err)
	}
}
