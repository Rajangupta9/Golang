package middleware

import (
	"GoBackend/utils"
	"context"
	"net/http"
	"strings"
)

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			utils.ResponseWithError(w, http.StatusUnauthorized, "Authorization header required")
			return
		}
		tokenString := strings.Replace(authHeader, "Bearer ", "", 1)
		// fmt.Println(tokenString)
		userID, err := utils.ValidateJwt(tokenString)

		if err != nil {
			utils.ResponseWithError(w, http.StatusUnauthorized, "Invalid Token")
			return
		}
		ctx := context.WithValue(r.Context(), "user_id", userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}
