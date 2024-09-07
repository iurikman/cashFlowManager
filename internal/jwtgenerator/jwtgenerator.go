package jwtgenerator

import (
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/iurikman/cashFlowManager/internal/models"
	log "github.com/sirupsen/logrus"
)

const (
	readerBits = 2048
	validDays  = 24 * time.Hour
)

type JWTGenerator struct {
	publicKey  *rsa.PublicKey
	privateKey *rsa.PrivateKey
}

func NewJWTGenerator() *JWTGenerator {
	privateKey, err := rsa.GenerateKey(rand.Reader, readerBits)
	if err != nil {
		log.Warnf("rsa.GenerateKey err: %v", err)
	}

	generator := &JWTGenerator{
		publicKey:  &privateKey.PublicKey,
		privateKey: privateKey,
	}

	return generator
}

func (j *JWTGenerator) GetNewTokenString(user models.User) (string, error) {
	claims := models.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "localhost:8080",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(validDays)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		UUID: user.ID,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS512, claims)

	ss, err := token.SignedString(j.privateKey)
	if err != nil {
		return "", fmt.Errorf("token.SignedString(j.privateKey) err: %w", err)
	}

	return ss, nil
}

func (j *JWTGenerator) GetPublicKey() *rsa.PublicKey {
	key := *j.publicKey

	return &key
}
