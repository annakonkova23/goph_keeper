// Package auth предоставляет gRPC-интерцепторы для проверки JWT-токенов.
//
// Интерцепторы используются для защиты приватных методов, позволяя доступ
// только авторизованным пользователям. Публичные методы (например, регистрация и вход)
// исключаются из проверки.
//
// Поддерживает:
//   - Универсальные (unary) и потоковые (stream) RPC.
//   - Извлечение userID из токена и сохранение его в контексте.
//   - Проверку заголовка "authorization: Bearer <token>".
//
// Пример использования:
//
//	jwtManager := auth.NewJWTManager("secret", 24*time.Hour)
//	interceptor := auth.UnaryAuthInterceptor(jwtManager)
//
//	grpcServer := grpc.NewServer(
//	    grpc.UnaryInterceptor(interceptor),
//	)
//
// В защищённых хендлерах можно получить userID:
//
//	userID, ok := ctx.Value(auth.UserIDKey).(string)
//	if !ok {
//	    return status.Error(codes.Unauthenticated, "пользователь не авторизован")
//	}
package auth

import (
	"context"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type contextKey string

// UserIDKey — ключ для хранения userID в контексте.
// Используется в обработчиках для извлечения идентификатора пользователя.
//
// Пример:
//
//	userID, ok := ctx.Value(auth.UserIDKey).(string)
const UserIDKey contextKey = "userID"

// UnaryAuthInterceptor возвращает gRPC-интерцептор для unary-запросов,
// который проверяет наличие и валидность JWT-токена в заголовке authorization.
//
// Публичные методы (например, /api.AuthService/Login) пропускаются без проверки.
//
// Этапы работы:
//  1. Проверяет, является ли метод публичным.
//  2. Извлекает metadata из контекста.
//  3. Парсит заголовок "authorization: Bearer <token>".
//  4. Проверяет токен с помощью jwtManager.Verify.
//  5. Добавляет userID в контекст.
//
// Возвращает:
//   - codes.Unauthenticated, если:
//   - нет metadata,
//   - нет токена,
//   - токен невалиден.
//
// Пример заголовка:
//
//	authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
func UnaryAuthInterceptor(jwtManager *JWTManager) grpc.UnaryServerInterceptor {
	publicMethods := map[string]bool{
		"/api.AuthService/Register": true,
		"/api.AuthService/Login":    true,
	}

	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		if publicMethods[info.FullMethod] {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "metadata не задан")
		}

		values := md.Get("authorization")
		if len(values) == 0 {
			return nil, status.Error(codes.Unauthenticated, "authorization token не задан")
		}

		parts := strings.SplitN(values[0], " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			return nil, status.Error(codes.Unauthenticated, "невалидный authorization header")
		}

		userID, err := jwtManager.Verify(parts[1])
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "невалидный token")
		}

		ctx = context.WithValue(ctx, UserIDKey, userID)
		return handler(ctx, req)
	}
}

// StreamAuthInterceptor возвращает gRPC-интерцептор для stream-запросов.
// Аналогичен UnaryAuthInterceptor, но работает с потоковыми соединениями.
//
// Оборачивает ServerStream, чтобы внедрить обновлённый контекст с userID.
//
// Публичные методы пропускаются без проверки.
//
// Возвращает ошибку codes.Unauthenticated при:
//   - отсутствии metadata,
//   - отсутствии или неверном формате токена,
//   - невалидном JWT.
func StreamAuthInterceptor(jwtManager *JWTManager) grpc.StreamServerInterceptor {
	publicMethods := map[string]bool{
		"/api.AuthService/Register": true,
		"/api.AuthService/Login":    true,
	}

	return func(
		srv any,
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {

		if publicMethods[info.FullMethod] {
			return handler(srv, ss)
		}

		md, ok := metadata.FromIncomingContext(ss.Context())
		if !ok {
			return status.Error(codes.Unauthenticated, "metadata не задан")
		}

		values := md.Get("authorization")
		if len(values) == 0 {
			return status.Error(codes.Unauthenticated, "authorization token не задан")
		}

		parts := strings.SplitN(values[0], " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			return status.Error(codes.Unauthenticated, "невалидный authorization header")
		}

		token := parts[1]
		userID, err := jwtManager.Verify(token)
		if err != nil {
			return status.Error(codes.Unauthenticated, "невалидный token")
		}

		ctx := context.WithValue(ss.Context(), UserIDKey, userID)

		wrapped := &serverStreamWrapper{
			ServerStream: ss,
			ctx:          ctx,
		}

		return handler(srv, wrapped)
	}
}

// serverStreamWrapper оборачивает grpc.ServerStream,
// чтобы заменить контекст с внедрённым userID.
//
// Используется в StreamAuthInterceptor для передачи
// аутентифицированного контекста в обработчик.
type serverStreamWrapper struct {
	grpc.ServerStream
	ctx context.Context
}

// Context возвращает контекст с внедрённым userID.
// Переопределяет стандартный метод ServerStream.Context().
func (w *serverStreamWrapper) Context() context.Context {
	return w.ctx
}
