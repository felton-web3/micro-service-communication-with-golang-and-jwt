package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"felton.com/microservicecomm/auth"
	"felton.com/microservicecomm/config"
	"felton.com/microservicecomm/middleware"
	fhttp "felton.com/microservicecomm/transport"
	"github.com/gorilla/mux"
)

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	fhttp.ResponseWithJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func publicProductsHandler(w http.ResponseWriter, r *http.Request) {
	products := []string{"Book", "Pen"}
	fhttp.ResponseWithJSON(w, http.StatusOK, products)
}

func privateProductsHandler(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		fhttp.ResponseWithError(w, http.StatusUnauthorized, "Invalid claims")
		slog.Info("Invalid claims")
		return
	}
	slog.Info("Accessing private products", "user_id", claims.UserID)
	products := []string{"Laptop (Private)", "Monitor (Private)"}

	//do sth, and response
	//or create a kafka message

	fhttp.ResponseWithJSON(w, http.StatusOK, products)
}

func adminDeleteProductHandler(w http.ResponseWriter, r *http.Request) {
	claims, _ := middleware.GetClaims(r.Context())

	// Role-Based Access Control (RBAC)
	isAdmin := false
	for _, role := range claims.Roles {
		if role == "admin" {
			isAdmin = true
			break
		}
	}

	if !isAdmin {
		fhttp.ResponseWithError(w, http.StatusForbidden, "Forbidden: admin role required")
		return
	}

	vars := mux.Vars(r)
	productID := vars["id"]

	slog.Info("Admin deleted a product", "user_id", claims.UserID, "product_id", productID)
	fhttp.ResponseWithJSON(w, http.StatusOK, map[string]string{"message": fmt.Sprintf("Product %s deleted by admin %s", productID, claims.UserID)})
}

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	config.Load()

	// 从认证服务获取公钥
	publicKey := auth.FetchPublicKeyFromURL(config.AppConfig.JWT.AuthServerPublicKeyURL)
	tokenValidator := auth.NewTokenValidator(publicKey)

	r := mux.NewRouter()
	r.Use(middleware.Logger, middleware.Recovery)

	// 公开路由
	r.HandleFunc("/health", healthCheckHandler).Methods("GET")
	r.HandleFunc("/api/products/public", publicProductsHandler).Methods("GET")

	// 受保护的路由组
	api := r.PathPrefix("/api").Subrouter()
	api.Use(middleware.Auth(tokenValidator))
	api.HandleFunc("/products/private", privateProductsHandler).Methods("GET")
	api.HandleFunc("/products/{id}", adminDeleteProductHandler).Methods("DELETE")

	//api.HandleFunc("/products/private", privateProductsHandler).Methods("GET")
	//比如 billing-service监听某个端口，当其他service 发送数据校验合格后，调用 billing-service内部处理逻辑。
	//其他 service 是否需要 先 login，获得 token，然后 后期发送数据，都需要这个token 携带进行校验

	// 注意：这里的端口应该和 auth-service 不同
	addr := ":8081"
	slog.Info("Product Service starting", "address", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		slog.Error("failed to start server", "error", err)
		os.Exit(1)
	}
}
