package auth

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"felton.com/microservicecomm/config"
	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID string   `json:"user_id"`
	Roles  []string `json: "roles"`
	jwt.RegisteredClaims
}

type TokenGenerator struct {
	privateKey *rsa.PrivateKey
	issuer     string
	accessDur  time.Duration
}

func NewTokenGenerator(privateKey *rsa.PrivateKey, config *config.Config) *TokenGenerator {
	return &TokenGenerator{
		privateKey: privateKey,
		issuer:     config.JWT.Issuer,
		accessDur:  config.JWT.AccessTokenDuration,
	}
}

func (tg *TokenGenerator) Generate(userID string, roles []string) (string, error) {
	claims := Claims{
		UserID: userID,
		Roles:  roles,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tg.accessDur)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    tg.issuer,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(tg.privateKey)
}

func (tg *TokenGenerator) Validate(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return tg.privateKey.Public(), nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}
	return nil, errors.New("invalid token")
}

type TokenValidator struct {
	publicKey *rsa.PublicKey
}

func NewTokenValidator(publicKey *rsa.PublicKey) *TokenValidator {
	return &TokenValidator{
		publicKey: publicKey,
	}
}

func (tv *TokenValidator) Validate(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return tv.publicKey, nil
	})

	slog.Info("token", "token", token)
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}
