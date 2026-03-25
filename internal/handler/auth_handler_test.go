package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/konkovaanna23/gophkeeper/internal/handler/middleware"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type mockAuthClient struct {
	registerFn func(ctx context.Context, login, password string) (string, error)
	loginFn    func(ctx context.Context, login, password string) (string, error)

	registerCalled bool
	loginCalled    bool

	gotRegisterLogin    string
	gotRegisterPassword string
	gotLoginLogin       string
	gotLoginPassword    string
}

func (m *mockAuthClient) Register(ctx context.Context, login, password string) (string, error) {
	m.registerCalled = true
	m.gotRegisterLogin = login
	m.gotRegisterPassword = password

	if m.registerFn != nil {
		return m.registerFn(ctx, login, password)
	}
	return "", nil
}

func (m *mockAuthClient) Login(ctx context.Context, login, password string) (string, error) {
	m.loginCalled = true
	m.gotLoginLogin = login
	m.gotLoginPassword = password

	if m.loginFn != nil {
		return m.loginFn(ctx, login, password)
	}
	return "", nil
}

func (m *mockAuthClient) Close() error {
	return nil
}

func TestServer_Register_Success(t *testing.T) {
	auth := &mockAuthClient{
		registerFn: func(ctx context.Context, login, password string) (string, error) {
			return "user-123", nil
		},
	}

	s := &Server{
		authClient: auth,
	}

	req := httptest.NewRequest(http.MethodPost, "/api/register", strings.NewReader(`{
		"login":"anna",
		"password":"secret"
	}`))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	s.Register(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("want status %d, got %d", http.StatusOK, rec.Code)
	}

	if !auth.registerCalled {
		t.Fatal("expected authClient.Register to be called")
	}

	if auth.gotRegisterLogin != "anna" {
		t.Fatalf("want login anna, got %q", auth.gotRegisterLogin)
	}

	if auth.gotRegisterPassword != "secret" {
		t.Fatalf("want password secret, got %q", auth.gotRegisterPassword)
	}

	var resp RegisterResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}

	if resp.UserID != "user-123" {
		t.Fatalf("want user_id user-123, got %q", resp.UserID)
	}
}

func TestServer_Register_BadJSON(t *testing.T) {
	auth := &mockAuthClient{}
	s := &Server{
		authClient: auth,
	}

	req := httptest.NewRequest(http.MethodPost, "/api/register", strings.NewReader(`{"login":"anna"`))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	s.Register(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("want status %d, got %d", http.StatusBadRequest, rec.Code)
	}

	if auth.registerCalled {
		t.Fatal("authClient.Register must not be called on invalid JSON")
	}
}

func TestServer_Register_GRPCError(t *testing.T) {
	auth := &mockAuthClient{
		registerFn: func(ctx context.Context, login, password string) (string, error) {
			return "", status.Error(codes.AlreadyExists, "user already exists")
		},
	}

	s := &Server{
		authClient: auth,
	}

	req := httptest.NewRequest(http.MethodPost, "/api/register", strings.NewReader(`{
		"login":"anna",
		"password":"secret"
	}`))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	s.Register(rec, req)

	if !auth.registerCalled {
		t.Fatal("expected authClient.Register to be called")
	}

	if rec.Code < 400 {
		t.Fatalf("want error status, got %d", rec.Code)
	}
}

func TestServer_Login_Success(t *testing.T) {
	auth := &mockAuthClient{
		loginFn: func(ctx context.Context, login, password string) (string, error) {
			return "token-123", nil
		},
	}

	s := &Server{
		authClient: auth,
	}

	req := httptest.NewRequest(http.MethodPost, "/api/login", strings.NewReader(`{
		"login":"anna",
		"password":"secret"
	}`))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	s.Login(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("want status %d, got %d", http.StatusOK, rec.Code)
	}

	if !auth.loginCalled {
		t.Fatal("expected authClient.Login to be called")
	}

	if auth.gotLoginLogin != "anna" {
		t.Fatalf("want login anna, got %q", auth.gotLoginLogin)
	}

	if auth.gotLoginPassword != "secret" {
		t.Fatalf("want password secret, got %q", auth.gotLoginPassword)
	}
}

func TestServer_Login_BadJSON(t *testing.T) {
	auth := &mockAuthClient{}
	s := &Server{
		authClient: auth,
	}

	req := httptest.NewRequest(http.MethodPost, "/api/login", strings.NewReader(`{"login":"anna"`))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	s.Login(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("want status %d, got %d", http.StatusBadRequest, rec.Code)
	}

	if auth.loginCalled {
		t.Fatal("authClient.Login must not be called on invalid JSON")
	}
}

func TestServer_Login_GRPCError(t *testing.T) {
	auth := &mockAuthClient{
		loginFn: func(ctx context.Context, login, password string) (string, error) {
			return "", status.Error(codes.Unauthenticated, "invalid credentials")
		},
	}

	s := &Server{
		authClient: auth,
	}

	req := httptest.NewRequest(http.MethodPost, "/api/login", strings.NewReader(`{
		"login":"anna",
		"password":"wrong"
	}`))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	s.Login(rec, req)

	if !auth.loginCalled {
		t.Fatal("expected authClient.Login to be called")
	}

	if rec.Code < 400 {
		t.Fatalf("want error status, got %d", rec.Code)
	}
}

func TestServer_Login_WithAuthCookieMiddleware_SetsCookie(t *testing.T) {
	auth := &mockAuthClient{
		loginFn: func(ctx context.Context, login, password string) (string, error) {
			return "token-123", nil
		},
	}

	s := &Server{
		authClient: auth,
	}

	const cookieName = "access_token"

	r := chi.NewRouter()
	r.Group(func(pr chi.Router) {
		pr.Use(middleware.WithAuthCookie(cookieName))
		pr.Post("/api/login", s.Login)
	})

	req := httptest.NewRequest(http.MethodPost, "/api/login", strings.NewReader(`{
		"login":"anna",
		"password":"secret"
	}`))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("want status %d, got %d", http.StatusOK, rec.Code)
	}

	if !auth.loginCalled {
		t.Fatal("expected authClient.Login to be called")
	}

	var gotCookie *http.Cookie
	for _, c := range rec.Result().Cookies() {
		if c.Name == cookieName {
			gotCookie = c
			break
		}
	}

	if gotCookie == nil {
		t.Fatalf("expected cookie %q to be set", cookieName)
	}

	if gotCookie.Value != "token-123" {
		t.Fatalf("want cookie value %q, got %q", "token-123", gotCookie.Value)
	}

	if !gotCookie.HttpOnly {
		t.Fatal("expected HttpOnly=true")
	}

	if !gotCookie.Secure {
		t.Fatal("expected Secure=true")
	}

	if gotCookie.Path != "/" {
		t.Fatalf("want cookie path %q, got %q", "/", gotCookie.Path)
	}
}
