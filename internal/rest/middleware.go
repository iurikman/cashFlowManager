package rest

import (
	"context"
	"crypto/rsa"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/iurikman/cashFlowManager/internal/models"
	log "github.com/sirupsen/logrus"
)

const headerLength = 2

func (s *Server) jwtAuth(next http.Handler) http.Handler {
	var fn http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
		claims, err := getClaimsFromHeader(r.Header.Get("Authorization"), s.key)
		switch {
		case errors.Is(err, models.ErrInvalidAccessToken):
			writeErrorResponse(w, http.StatusUnauthorized, "invalid access token")

			return
		case err != nil:
			log.Warnf("getClaimsFromHeader err: %v", err)
			writeErrorResponse(w, http.StatusInternalServerError, "internal server error")

			return
		}

		espiresAtTime := time.Unix(claims.ExpiresAt.Unix(), 0)
		if espiresAtTime.Before(time.Now()) {
			writeErrorResponse(w, http.StatusUnauthorized, "invalid access token")

			return
		}

		userInfo := models.UserInfo{
			ID: claims.UUID,
		}

		r = r.WithContext(context.WithValue(r.Context(), models.UserInfoKey, userInfo))
		next.ServeHTTP(w, r)
	}

	return fn
}

func getClaimsFromHeader(authHeader string, key *rsa.PublicKey) (*models.Claims, error) {
	if authHeader == "" {
		return nil, models.ErrHeaderIsEmpty
	}

	headerParts := strings.Split(authHeader, " ")

	if len(headerParts) != headerLength || headerParts[0] != "Bearer" {
		return nil, models.ErrInvalidAccessToken
	}

	claims, err := parseToken(headerParts[1], key)

	switch {
	case errors.Is(err, models.ErrInvalidAccessToken):
		return nil, models.ErrInvalidAccessToken
	case err != nil:
		return nil, fmt.Errorf("jwt.parseToken(headerParts[1], key) err: %w", err)
	}

	return claims, nil
}

func parseToken(accessToken string, key *rsa.PublicKey) (*models.Claims, error) {
	token, err := jwt.ParseWithClaims(accessToken, &models.Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, models.ErrInvalidAccessToken
		}

		return key, nil
	})
	if errors.Is(err, jwt.ErrTokenMalformed) {
		return nil, models.ErrInvalidAccessToken
	}

	if err != nil {
		return nil, fmt.Errorf("jwt.ParseWithClaims err: %w", err)
	}

	claims, ok := token.Claims.(*models.Claims)
	if !ok {
		return nil, models.ErrInvalidAccessToken
	}

	return claims, nil
}

func (s *Server) getOwnerIDFromRequest(r *http.Request) uuid.UUID {
	userInfo, _ := r.Context().Value(models.UserInfoKey).(models.UserInfo)

	return userInfo.ID
}
