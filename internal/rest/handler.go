package rest

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/iurikman/cashFlowManager/internal/models"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

type service interface {
	CreateWallet(context context.Context, wallet models.Wallet) (*models.Wallet, error)
	GetWalletByID(ctx context.Context, id uuid.UUID) (*models.Wallet, error)
	UpdateWallet(context context.Context, id uuid.UUID, wallet models.WalletDTO) (*models.Wallet, error)
	DeleteWallet(context context.Context, id uuid.UUID) error
	Deposit(ctx context.Context, transaction models.Transaction) error
	Transfer(ctx context.Context, transaction models.Transaction, initAmount float64) error
	Withdraw(ctx context.Context, transaction models.Transaction) error
}

type HTTPResponse struct {
	Data  any    `json:"data"`
	Error string `json:"error"`
}

func (s *Server) createWallet(w http.ResponseWriter, r *http.Request) {
	var wallet models.Wallet

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewDecoder(r.Body).Decode(&wallet); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, err.Error())

		return
	}

	if err := wallet.Validate(); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, err.Error())

		return
	}

	createdUser, err := s.service.CreateWallet(r.Context(), wallet)
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "internal server error")

		return
	}

	writeOkResponse(w, http.StatusCreated, createdUser)
}

func (s *Server) getWalletByID(w http.ResponseWriter, r *http.Request) {
	walletID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeErrorResponse(w, http.StatusBadRequest, err.Error())

		return
	}

	wallet, err := s.service.GetWalletByID(r.Context(), walletID)

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
	var walletDTO models.WalletDTO

	walletID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeErrorResponse(w, http.StatusBadRequest, err.Error())

		return
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewDecoder(r.Body).Decode(&walletDTO); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, err.Error())

		return
	}

	updatedWallet, err := s.service.UpdateWallet(r.Context(), walletID, walletDTO)

	switch {
	case errors.Is(err, models.ErrWalletNotFound):
		writeErrorResponse(w, http.StatusNotFound, err.Error())

		return
	case err != nil:
		writeErrorResponse(w, http.StatusInternalServerError, "internal server error")

		return
	}

	writeOkResponse(w, http.StatusOK, updatedWallet)
}

func (s *Server) deleteWallet(w http.ResponseWriter, r *http.Request) {
	walletID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeErrorResponse(w, http.StatusBadRequest, err.Error())

		return
	}

	err = s.service.DeleteWallet(r.Context(), walletID)

	switch {
	case errors.Is(err, models.ErrWalletNotFound):
		writeErrorResponse(w, http.StatusNotFound, err.Error())

		return
	case err != nil:
		writeErrorResponse(w, http.StatusInternalServerError, "internal server error")

		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) deposit(w http.ResponseWriter, r *http.Request) {
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

	err := s.service.Deposit(r.Context(), transaction)

	switch {
	case errors.Is(err, models.ErrWalletNotFound):
		writeErrorResponse(w, http.StatusNotFound, err.Error())

		return
	case err != nil:
		writeErrorResponse(w, http.StatusInternalServerError, "internal server error")

		return
	}

	writeOkResponse(w, http.StatusOK, nil)
}

func (s *Server) transfer(w http.ResponseWriter, r *http.Request) {
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

	err := s.service.Transfer(r.Context(), transaction, transaction.Amount)

	switch {
	case errors.Is(err, models.ErrWalletNotFound):
		writeErrorResponse(w, http.StatusNotFound, err.Error())

		return
	case err != nil:
		writeErrorResponse(w, http.StatusInternalServerError, "internal server error")

		return
	}

	writeOkResponse(w, http.StatusOK, nil)
}

func (s *Server) withdraw(w http.ResponseWriter, r *http.Request) {
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

	err := s.service.Withdraw(r.Context(), transaction)

	switch {
	case errors.Is(err, models.ErrWalletNotFound):
		writeErrorResponse(w, http.StatusNotFound, err.Error())

		return
	case err != nil:
		writeErrorResponse(w, http.StatusInternalServerError, "internal server error")

		return
	}

	writeOkResponse(w, http.StatusOK, nil)
}

func writeOkResponse(w http.ResponseWriter, statusCode int, respData any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(HTTPResponse{Data: respData}); err != nil {
		log.Warnf("json.NewEncoder(w).Encode(HTTPResponse{Data: respData}) err: %v", err)
	}
}

func writeErrorResponse(w http.ResponseWriter, statusCode int, description string) {
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(HTTPResponse{Error: description}); err != nil {
		log.Warnf("json.NewEncoder(w).Encode(HTTPResponse{Error: description}) err: %s", err)
	}
}
