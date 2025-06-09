package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"felton.com/microservicecomm/auth"
	"felton.com/microservicecomm/config"
	"felton.com/microservicecomm/middleware"
	fhttp "felton.com/microservicecomm/transport"
	"github.com/gorilla/mux"
)

var tokenGenerator *auth.TokenGenerator

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// 模拟用户数据库
var users = map[string]struct {
	Password string
	Roles    []string
}{
	"admin": {"password123", []string{"admin", "user"}},
	"user1": {"password123", []string{"user"}},
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		fhttp.ResponseWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	user, ok := users[req.Username]
	if !ok || user.Password != req.Password {
		fhttp.ResponseWithError(w, http.StatusUnauthorized, "Invalid username or password")
		return
	}

	token, err := tokenGenerator.Generate(req.Username, user.Roles)
	if err != nil {
		fhttp.ResponseWithError(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	// 检查 token 是否已过期，如果过期则重新生成
	claims, err := tokenGenerator.Validate(token)
	if err != nil || claims.ExpiresAt.Before(time.Now()) {
		token, err = tokenGenerator.Generate(req.Username, user.Roles)
		if err != nil {
			fhttp.ResponseWithError(w, http.StatusInternalServerError, "Failed to regenerate token")
			return
		}
	}

	fhttp.ResponseWithJSON(w, http.StatusOK, map[string]string{"access_token": token})
}

func publicKeyHandler(w http.ResponseWriter, r *http.Request) {
	keyData, err := os.ReadFile(config.AppConfig.JWT.PublicKeyFile)
	if err != nil {
		fhttp.ResponseWithError(w, http.StatusInternalServerError, "Could not read public key")
		return
	}
	w.Header().Set("Content-Type", "application/x-pem-file")
	w.Write(keyData)
}

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	config.Load()

	privateKey := auth.LoadPrivateKey(config.AppConfig.JWT.PrivateKeyFile)

	tokenGenerator = auth.NewTokenGenerator(privateKey, config.AppConfig)

	r := mux.NewRouter()
	r.Use(middleware.Logger, middleware.Recovery)

	r.HandleFunc("/api/login", loginHandler).Methods("POST")
	r.HandleFunc("/api/public-key", publicKeyHandler).Methods("GET")

	addr := fmt.Sprintf(":%d", config.AppConfig.Server.Port)
	slog.Info("Auth Service starting", "address", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		slog.Error("failed to start server", "error", err)
		os.Exit(1)
	}
}
