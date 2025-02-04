package middlewares

import (
	"net/http"
	"strings"

	"github.com/anuntech/finance-backend/internal/utils"
)

func VerifyAccessToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authorization := r.Header.Get("Authorization")

		if authorization == "" {
			http.Error(w, "Missing or invalid access token", http.StatusUnauthorized)
			return
		}

		authorization = strings.TrimPrefix(authorization, "Bearer ")

		claims, err := utils.NewCreateAccessTokenUtil().DecodeToken(authorization)

		if err != nil {
			http.Error(w, "Invalid or expired access token", http.StatusUnauthorized)
			return
		}

		r.Header.Set("UserId", claims["sub"].(string))

		next.ServeHTTP(w, r)
	})
}
