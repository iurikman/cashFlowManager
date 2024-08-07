package rest

import (
	"fmt"
	"net/http"
)

type service interface{}

//nolint:forbidigo
func (s *Server) createWallet(w http.ResponseWriter, r *http.Request) {
	wr := w
	rr := r

	fmt.Print(wr)
	fmt.Print(rr)
}
