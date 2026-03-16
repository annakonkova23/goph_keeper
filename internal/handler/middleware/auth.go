// Package middleware
package middleware

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"net/http"
	"strings"
)

type ctxKey struct{ key string }

var userIDKey = ctxKey{"user_id"}

const cookieName = "user"

// GetUserID извлекает userID из контекста.
func GetUserID(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(userIDKey).(string)
	return userID, ok
}

// SetUserID устанавливает userID в контекст.
func SetUserID(r *http.Request, userID string) *http.Request {
	ctx := context.WithValue(r.Context(), userIDKey, userID)
	return r.WithContext(ctx)
}

func sign(key string, loadString string) string {
	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(loadString))
	sum := h.Sum(nil)
	return base64.RawURLEncoding.EncodeToString(sum)
}

func verify(key, loadString, sig string) bool {
	expected := sign(key, loadString)
	return hmac.Equal([]byte(expected), []byte(sig))
}

// GenerateUserID генерирует случайный userID.
func GenerateUserID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(b[:]), nil
}

func parseCookieValue(v string) (userID, sig string, err error) {
	parts := strings.Split(v, ".")
	if len(parts) != 2 {
		return "", "", errors.New("некорректный формат cookie")
	}
	return parts[0], parts[1], nil
}

// AuthMiddleware - middleware для аутентификации пользователей.
func AuthMiddleware(key string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			var userID string
			var needSetCookie bool

			c, err := r.Cookie(cookieName)
			if err == nil {
				id, sig, err := parseCookieValue(c.Value)
				if err != nil || id == "" {
					http.Error(w,
						http.StatusText(http.StatusUnauthorized),
						http.StatusUnauthorized,
					)
					return
				}

				if verify(key, id, sig) {
					userID = id
				} else {
					needSetCookie = true
				}
			} else {
				needSetCookie = true
			}

			if needSetCookie {
				id, err := GenerateUserID()
				if err != nil {
					http.Error(w,
						http.StatusText(http.StatusInternalServerError),
						http.StatusInternalServerError,
					)
					return
				}
				userID = id
				sig := sign(key, userID)
				value := userID + "." + sig

				http.SetCookie(w, &http.Cookie{
					Name:     cookieName,
					Value:    value,
					Path:     "/",
					HttpOnly: true,
					Secure:   false,
				})
			}

			r = SetUserID(r, userID)

			next.ServeHTTP(w, r)
		})
	}
}
