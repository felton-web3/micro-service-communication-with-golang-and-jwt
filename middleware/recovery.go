package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"

	response "felton.com/microservicecomm/transport"
)

func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				slog.Error("service panicked", "error", err, "stack", string(debug.Stack()))
				response.ResponseWithError(w, http.StatusInternalServerError, "Internal Server Error")
			}
		}()
		next.ServeHTTP(w, r)
	})
}
