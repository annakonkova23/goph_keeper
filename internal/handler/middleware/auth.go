package middleware

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/konkovaanna23/gophkeeper/internal/handler/helper"
	"google.golang.org/grpc/metadata"
)

type authCookieSinkKey struct{}

type authCookieSink struct {
	AccessToken string
}

type bufferedResponseWriter struct {
	header     http.Header
	statusCode int
	body       bytes.Buffer
}

func newBufferedResponseWriter() *bufferedResponseWriter {
	return &bufferedResponseWriter{
		header:     make(http.Header),
		statusCode: http.StatusOK,
	}
}

func (w *bufferedResponseWriter) Header() http.Header {
	return w.header
}

func (w *bufferedResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
}

func (w *bufferedResponseWriter) Write(p []byte) (int, error) {
	return w.body.Write(p)
}

func (w *bufferedResponseWriter) FlushTo(dst http.ResponseWriter) error {
	copyHeaders(dst.Header(), w.header)
	dst.WriteHeader(w.statusCode)

	_, err := dst.Write(w.body.Bytes())
	return err
}

func copyHeaders(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

func WithGRPCAuthorization(cookieName string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(cookieName)
			if err != nil {
				helper.WriteJSON(w, http.StatusUnauthorized, map[string]any{
					"ok":    false,
					"error": "missing auth cookie",
				})
				return
			}

			token := strings.TrimSpace(cookie.Value)
			if token == "" {
				helper.WriteJSON(w, http.StatusUnauthorized, map[string]any{
					"ok":    false,
					"error": "empty auth cookie",
				})
				return
			}

			fmt.Println("ЗАДАЁМ ТОКЕН")
			fmt.Println(token)

			ctx := metadata.AppendToOutgoingContext(
				r.Context(),
				"authorization", "Bearer "+token,
			)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func WithAuthCookie(cookieName string, cookieSecure bool, cookieTTL time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sink := &authCookieSink{}

			ctx := context.WithValue(r.Context(), authCookieSinkKey{}, sink)
			r = r.WithContext(ctx)

			// Буферизуем ответ, чтобы можно было безопасно добавить Set-Cookie
			// даже после того, как handler уже "записал" ответ.
			bw := newBufferedResponseWriter()
			next.ServeHTTP(bw, r)

			if sink.AccessToken != "" {
				http.SetCookie(bw, &http.Cookie{
					Name:     cookieName,
					Value:    sink.AccessToken,
					Path:     "/",
					HttpOnly: true,
					Secure:   cookieSecure,
					SameSite: http.SameSiteLaxMode,
					MaxAge:   int(cookieTTL.Seconds()),
					Expires:  time.Now().Add(cookieTTL),
				})
			}

			_ = bw.FlushTo(w)
		})
	}
}

func SetAccessToken(ctx context.Context, token string) bool {
	sink, ok := ctx.Value(authCookieSinkKey{}).(*authCookieSink)
	if !ok || sink == nil {
		return false
	}

	sink.AccessToken = token
	return true
}
