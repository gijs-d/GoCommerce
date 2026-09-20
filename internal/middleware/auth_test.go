package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"webshop/internal/auth"
	"webshop/internal/middleware"
)

type mockStore struct {
	users map[string]*auth.SessionUser
}

func (m *mockStore) GetUserBySession(ctx context.Context, token string) (*auth.SessionUser, error) {
	if u, ok := m.users[token]; ok {
		return u, nil
	}
	return nil, nil
}

func TestSessionMiddleware_NoToken(t *testing.T) {
	store := &mockStore{users: map[string]*auth.SessionUser{}}
	handler := middleware.SessionMiddleware(store, "")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/user/profile", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Verwachtte status 401 Unauthorized, kreeg %d", rr.Code)
	}
}

func TestSessionMiddleware_ValidCustomer(t *testing.T) {
	token := "valid-customer-token-12345"
	store := &mockStore{
		users: map[string]*auth.SessionUser{
			token: {
				ID:       "user-1",
				Email:    "customer@example.com",
				FullName: "Gijs Klant",
				Role:     "customer",
			},
		},
	}

	var capturedUser *auth.SessionUser
	handler := middleware.SessionMiddleware(store, "customer")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedUser = middleware.GetUserFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/customer/orders", nil)
	req.AddCookie(&http.Cookie{Name: "session_token", Value: token})
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Verwachtte status 200 OK, kreeg %d", rr.Code)
	}

	if capturedUser == nil || capturedUser.Email != "customer@example.com" {
		t.Errorf("Gebruiker niet correct doorgegeven in context")
	}
}

func TestSessionMiddleware_RoleForbidden(t *testing.T) {
	token := "customer-trying-admin"
	store := &mockStore{
		users: map[string]*auth.SessionUser{
			token: {
				ID:   "user-2",
				Role: "customer",
			},
		},
	}

	handler := middleware.SessionMiddleware(store, "admin")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/admin/products", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("Verwachtte status 403 Forbidden voor gewone klant op admin route, kreeg %d", rr.Code)
	}
}