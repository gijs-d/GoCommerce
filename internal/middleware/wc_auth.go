package middleware

import (
	"context"
	"net/http"

	"webshop/internal/api"
	"webshop/internal/repository"
)

type wcAuthKey string

const (
	WcPermissionsKey wcAuthKey = "wc_permissions"
)

func WcAuthMiddleware(apiKeyRepo *repository.ApiKeyRepo, requireWrite bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var consumerKey, consumerSecret string

			// 1. Probeer HTTP Basic Auth (standaard bij WooCommerce clients)
			user, pass, ok := r.BasicAuth()
			if ok {
				consumerKey = user
				consumerSecret = pass
			} else {
				// 2. Probeer query parameters (?consumer_key=ck_...&consumer_secret=cs_...)
				q := r.URL.Query()
				consumerKey = q.Get("consumer_key")
				consumerSecret = q.Get("consumer_secret")
			}

			if consumerKey == "" || consumerSecret == "" {
				api.Error(w, http.StatusUnauthorized, "WooCommerce API authenticatie vereist (consumer_key & consumer_secret)")
				return
			}

			permissions, err := apiKeyRepo.ValidateKey(r.Context(), consumerKey, consumerSecret)
			if err != nil {
				api.Error(w, http.StatusUnauthorized, err.Error())
				return
			}

			if requireWrite && permissions != "write" && permissions != "read_write" {
				api.Error(w, http.StatusForbidden, "Onvoldoende rechten: schrijfrechten vereist")
				return
			}

			ctx := context.WithValue(r.Context(), WcPermissionsKey, permissions)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}