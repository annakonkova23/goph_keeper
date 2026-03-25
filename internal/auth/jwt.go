// Package auth предоставляет инструменты для генерации и проверки JWT-токенов.
//
// Основной компонент — JWTManager, который:
//   - Генерирует токены с указанным сроком жизни (TTL).
//   - Проверяет подпись и валидность токена.
//   - Извлекает userID из утверждений (claims).
//
// Поддерживает алгоритм HS256.
//
// Пример использования:
//
//	manager := auth.NewJWTManager("my-super-secret", 24*time.Hour)
//
//	// Генерация
//	token, err := manager.Generate("user-123")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Проверка
//	userID, err := manager.Verify(token)
//	if err != nil {
//	    log.Println("Неавторизован:", err)
//	} else {
//	    fmt.Println("Пользователь:", userID)
//	}
package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTManager отвечает за создание и проверку JWT-токенов.
//
// Содержит:
//   - secretKey: ключ для подписи токенов (должен быть надёжным).
//   - tokenTTL: время жизни токена (например, 24 * time.Hour).
type JWTManager struct {
	secretKey string
	tokenTTL  time.Duration
}

// UserClaims представляет пользовательские утверждения (claims) в JWT.
//
// Расширяет jwt.RegisteredClaims стандартными полями:
//   - IssuedAt, ExpiresAt, NotBefore, Subject.
//
// Дополнительно содержит:
//   - UserID: идентификатор пользователя (копия Subject).
//
// JSON-теги используются при сериализации/десериализации токена.
//
// Пример содержимого claims:
//
//	{
//	  "user_id": "user-123",
//	  "sub": "user-123",
//	  "iat": 1717000000,
//	  "nbf": 1717000000,
//	  "exp": 1717086400
//	}
type UserClaims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

// NewJWTManager создаёт новый менеджер JWT.
//
// Параметры:
//   - secretKey: секретный ключ для подписи (не должен передаваться в открытом виде).
//   - tokenTTL: длительность действия токена.
//
// Пример:
//
//	manager := auth.NewJWTManager("secret123", 1*time.Hour)
func NewJWTManager(secretKey string, tokenTTL time.Duration) *JWTManager {
	return &JWTManager{
		secretKey: secretKey,
		tokenTTL:  tokenTTL,
	}
}

// Generate создаёт JWT-токен для указанного пользователя.
//
// Включает в claims:
//   - user_id: идентификатор пользователя.
//   - subject: совпадает с user_id.
//   - issued_at: время выдачи.
//   - not_before: действителен с этого времени.
//   - expires_at: срок действия (now + tokenTTL).
//
// Возвращает:
//   - строку токена (в формате JWT).
//   - ошибку, если не удалось подписать.
//
// Пример:
//
//	token, err := manager.Generate("user-123")
//	if err != nil {
//	    return err
//	}
//	fmt.Println(token) // eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
func (m *JWTManager) Generate(userID string) (string, error) {
	now := time.Now()

	claims := UserClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.tokenTTL)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(m.secretKey))
}

// Verify проверяет подпись и срок действия JWT-токена.
//
// Проверяет:
//   - Корректность подписи (с использованием secretKey).
//   - Действителен ли токен на текущий момент (не истёк, не "до" времени).
//   - Наличие UserID в claims.
//
// Возвращает:
//   - userID, если токен валиден.
//   - ошибку, если:
//   - подпись неверна,
//   - срок действия истёк,
//   - токен повреждён,
//   - userID пустой.
//
// Пример:
//
//	userID, err := manager.Verify("eyJhbGci...")
//	if err != nil {
//	    return status.Error(codes.Unauthenticated, "невалидный токен")
//	}
func (m *JWTManager) Verify(tokenString string) (string, error) {
	claims := &UserClaims{}

	token, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(token *jwt.Token) (any, error) {
			return []byte(m.secretKey), nil
		},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
	)
	if err != nil {
		return "", err
	}

	if !token.Valid || claims.UserID == "" {
		return "", errors.New("невалидный token")
	}

	return claims.UserID, nil
}
