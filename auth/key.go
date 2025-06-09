package auth

import (
	"crypto/rsa"
	"io"
	"log/slog"
	"net/http"
	"os"

	"github.com/golang-jwt/jwt/v5"
)

func LoadPrivateKey(path string) *rsa.PrivateKey {
	keyData, err := os.ReadFile(path)
	if err != nil {
		slog.Error("failed to read private key file", "path", path, "error", err)
		os.Exit(1)
	}
	key, err := jwt.ParseRSAPrivateKeyFromPEM(keyData)

	if err != nil {
		slog.Error("failed to ParseRSAPrivateKeyFromPEM private key file", "path", path, "error", err)
		os.Exit(1)
	}
	return key
}

func LoadPublicKey(path string) *rsa.PublicKey {
	keyData, err := os.ReadFile(path)
	if err != nil {
		slog.Error("failed to read PublicKey key file", "path", path, "error", err)
		os.Exit(1)
	}

	key, err := jwt.ParseRSAPublicKeyFromPEM(keyData)
	if err != nil {
		slog.Error("failed to ParseRSAPublicKeyFromPEM private key file", "path", path, "error", err)
		os.Exit(1)
	}
	return key
}

func FetchPublicKeyFromURL(url string) *rsa.PublicKey {
	resp, err := http.Get(url)
	if err != nil {
		slog.Error("failed to fetch public key from URL", "url", url, "error", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	keyData, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("failed to ReadAll public key from URL", "url", url, "error", err)
		os.Exit(1)
	}
	key, err := jwt.ParseRSAPublicKeyFromPEM(keyData)
	if err != nil {
		slog.Error("failed to ParseRSAPublicKeyFromPEM public key from URL", "url", url, "error", err)
		os.Exit(1)
	}
	slog.Info("successfully fetched and parsed public key", "url", url)
	return key
}
