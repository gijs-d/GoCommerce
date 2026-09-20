package middleware

import (
	"context"
	"net/http"
	"strings"

	"webshop/internal/api"
	"webshop/internal/auth"
)

type contextKey string

const (
	UserContextKey    contextKey = "currentUser"
	SessionContextKey contextKey = "sessionToken"
)

// SessionStore interface zodat middleware ontkoppeld is en makkelijk te testen is
type SessionStore interface {
	GetUserBySession(ctx context.Context, token string) (*auth.SessionUser, error)
}

func SessionMiddleware(store SessionStore, requiredRole string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenString := ExtractSessionToken(r)
			if tokenString == "" {
				api.Error(w, http.StatusUnauthorized, "Inloggen vereist")
				return
			}

			user, err := store.GetUserBySession(r.Context(), tokenString)
			if err != nil || user == nil {
				api.Error(w, http.StatusUnauthorized, "Sessie verlopen of ongeldig")
				return
			}

			if requiredRole != "" && user.Role != requiredRole && user.Role != "admin" {
				api.Error(w, http.StatusForbidden, "Onvoldoende rechten voor deze bewerking")
				return
			}

			ctx := context.WithValue(r.Context(), UserContextKey, user)
			ctx = context.WithValue(ctx, SessionContextKey, tokenString)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func OptionalSession(store SessionStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenString := ExtractSessionToken(r)
			if tokenString != "" {
				if user, err := store.GetUserBySession(r.Context(), tokenString); err == nil && user != nil {
					ctx := context.WithValue(r.Context(), UserContextKey, user)
					ctx = context.WithValue(ctx, SessionContextKey, tokenString)
					next.ServeHTTP(w, r.WithContext(ctx))
					return
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

func GetUserFromContext(ctx context.Context) *auth.SessionUser {
	user, ok := ctx.Value(UserContextKey).(*auth.SessionUser)
	if !ok {
		return nil
	}
	return user
}

func GetSessionTokenFromContext(ctx context.Context) string {
	token, ok := ctx.Value(SessionContextKey).(string)
	if !ok {
		return ""
	}
	return token
}

func ExtractSessionToken(r *http.Request) string {
	// 1. Probeer HTTP-only sessiecookie
	if cookie, err := r.Cookie("session_token"); err == nil && cookie.Value != "" {
		return cookie.Value
	}

	// 2. Probeer Bearer token header
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		parts := strings.Split(authHeader, " ")
		if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
			return parts[1]
		}
	}

	return ""
}