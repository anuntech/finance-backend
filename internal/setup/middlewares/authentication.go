package middlewares

import (
	"net/http"
	"strings"

	"github.com/anuntech/finance-backend/internal/utils"
)

func VerifyAccessToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var authorization string
		if cookie, err := r.Cookie("__Secure-next-auth.session-token"); err == nil {
			authorization = cookie.Value
		} else if cookie, err := r.Cookie("next-auth.session-token"); err == nil {
			authorization = cookie.Value
		} else {
			http.Error(w, "Missing or invalid access token", http.StatusUnauthorized)
			return
		}

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
