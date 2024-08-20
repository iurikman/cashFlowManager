package rest

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/iurikman/cashFlowManager/internal/models"

	log "github.com/sirupsen/logrus"
)

type service interface {
	CreateWallet(context context.Context, wallet models.Wallet)
}

type HTTPResponse struct {
	Data  string `json:"data"`
	Error string `json:"error"`
}

func (s *Server) createWallet(w http.ResponseWriter, r *http.Request) {
	var wallet models.Wallet

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewDecoder(r.Body).Decode(&wallet); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "internal server error")

		return
	}

	s.service.CreateWallet(r.Context(), wallet)
}

func writeErrorResponse(w http.ResponseWriter, statusCode int, description string) {
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(HTTPResponse{Error: description}); err != nil {
		log.Panicf("json.NewEncoder(w).Encode(HTTPResponse{Error: description}) err: %s", err)

		return
	}
}
