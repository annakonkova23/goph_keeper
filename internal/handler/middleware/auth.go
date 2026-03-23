package middleware

import (
	"bytes"
	"context"
	"net/http"
	"strings"

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

// WithGRPCAuthorization создаёт middleware, который:
//   - Извлекает JWT-токен из указанной куки.
//   - Проверяет, что кука существует и не пустая.
//   - Добавляет токен в gRPC-метаданные как "authorization: Bearer <token>".
//   - Передаёт управление следующему обработчику.
//
// Если кука отсутствует или пустая — возвращает 401 Unauthorized.
//
// Используется для прозрачной передачи аутентификации
// от HTTP-слоя к gRPC-клиенту.
//
// Параметры:
//   - cookieName: имя куки (например, "auth").
//
// Пример:
//
//	mux := http.NewServeMux()
//	mux.Handle("/upload", WithGRPCAuthorization("auth")(uploadHandler))
//
// В uploadHandler токен будет доступен в контексте для gRPC-вызовов.
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

			ctx := metadata.AppendToOutgoingContext(
				r.Context(),
				"authorization", "Bearer "+token,
			)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// WithAuthCookie создаёт middleware, который:
//   - Устанавливает механизм записи access_token в куку.
//   - Перехватывает вызовы SetAccessToken(ctx, token) в цепочке обработчиков.
//   - После выполнения всех обработчиков — устанавливает куку, если токен был задан.
//
// Используется для того, чтобы после успешного логина
// автоматически отправить куку клиенту.
//
// Параметры:
//   - cookieName: имя куки (например, "auth").
//   - cookieSecure: если true — кука будет отправляться только по HTTPS.
//   - cookieTTL: время жизни куки (например, 24 * time.Hour).
//
// Кука устанавливается с параметрами:
//   - HttpOnly: true (защита от XSS)
//   - SameSite: Lax
//   - Path: "/"
//   - MaxAge и Expires: на основе cookieTTL
//
// Пример:
//
//	router.Use(WithAuthCookie("auth", true, 24*time.Hour))
//
//	func loginHandler(w http.ResponseWriter, r *http.Request) {
//	    // ... аутентификация
//	    middleware.SetAccessToken(r.Context(), "new-jwt-token")
//	    helper.WriteJSON(w, http.StatusOK, map[string]any{"ok": true})
//	}
//
// → Клиент получит куку "auth" с токеном.
func WithAuthCookie(cookieName string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sink := &authCookieSink{}

			ctx := context.WithValue(r.Context(), authCookieSinkKey{}, sink)
			r = r.WithContext(ctx)

			bw := newBufferedResponseWriter()
			next.ServeHTTP(bw, r)

			if sink.AccessToken != "" {
				http.SetCookie(bw, &http.Cookie{
					Name:     cookieName,
					Value:    sink.AccessToken,
					Path:     "/",
					HttpOnly: true,
					Secure:   false,
				})
			}

			_ = bw.FlushTo(w)
		})
	}
}

// SetAccessToken сохраняет access_token в контексте запроса
// для последующей установки через WithAuthCookie.
//
// Используется в обработчиках (например, при логине),
// чтобы сообщить middleware о необходимости установить куку.
//
// Возвращает:
//   - true: если sink найден и токен установлен.
//   - false: если контекст не содержит sink (например, middleware не подключён).
//
// Пример:
//
//	if valid {
//	    middleware.SetAccessToken(r.Context(), token)
//	    helper.WriteJSON(w, http.StatusOK, map[string]any{"ok": true})
//	}
//
// → WithAuthCookie перехватит этот токен и установит куку.
func SetAccessToken(ctx context.Context, token string) bool {
	sink, ok := ctx.Value(authCookieSinkKey{}).(*authCookieSink)
	if !ok || sink == nil {
		return false
	}

	sink.AccessToken = token
	return true
}
