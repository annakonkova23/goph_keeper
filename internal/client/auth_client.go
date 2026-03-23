// Package client предоставляет gRPC-клиент для взаимодействия с GophKeeper.
//
// Основной компонент — AuthServiceClient — позволяет:
//   - Регистрировать новых пользователей.
//   - Выполнять вход (получать JWT-токен).
//
// Все вызовы происходят через безопасное gRPC-соединение.
// Поддерживается работа с зашифрованными данными на стороне клиента.
//
// Пример использования:
//
//	client, err := client.NewAuthServiceClient("localhost:50051")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer client.Close()
//
//	userID, err := client.Register("alice@example.com", "secret")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	token, err := client.Login("alice@example.com", "secret")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	fmt.Println("Token:", token)
package client

import (
	"context"

	pb "github.com/konkovaanna23/gophkeeper/pkg/keeperservice"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// AuthClient — интерфейс для работы с аутентификацией.
// Полезен для мокирования в тестах.
//
// Методы:
//   - Register: создаёт нового пользователя.
//   - Login: возвращает access_token при успешной аутентификации.
//   - Close: закрывает соединение с сервером.
type AuthClient interface {
	Register(ctx context.Context, login, password string) (string, error)
	Login(ctx context.Context, login, password string) (string, error)
	Close() error
}

// AuthServiceClient — реализация клиента для gRPC-сервиса аутентификации.
//
// Использует сединение gRPC и генерированный pb.AuthServiceClient
// для вызова методов Register и Login.
type AuthServiceClient struct {
	conn   *grpc.ClientConn
	client pb.AuthServiceClient
}

// NewAuthServiceClient создаёт новый клиент для AuthService.
//
// Подключается к серверу по указанному адресу (например, "localhost:50051").
// Использует незащищённое соединение (insecure) — подходит для разработки.
// В продакшене рекомендуется использовать TLS.
//
// Возвращает:
//   - *AuthServiceClient: готовый к использованию клиент.
//   - ошибка, если не удалось установить соединение.
//
// Пример:
//
//	client, err := client.NewAuthServiceClient("localhost:50051")
//	if err != nil {
//	    log.Fatal("Не удалось подключиться:", err)
//	}
//	defer client.Close()
func NewAuthServiceClient(grpcAddr string) (*AuthServiceClient, error) {
	conn, err := grpc.NewClient(
		grpcAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}

	return &AuthServiceClient{
		conn:   conn,
		client: pb.NewAuthServiceClient(conn),
	}, nil
}

// Close закрывает gRPC-соединение.
//
// Безопасно вызывать на nil-указатель или если conn == nil.
// Может вызываться несколько раз — повторные вызовы игнорируются.
//
// Должен вызываться в defer после создания клиента.
//
// Пример:
//
//	defer client.Close()
func (c *AuthServiceClient) Close() error {
	if c == nil || c.conn == nil {
		return nil
	}
	return c.conn.Close()
}

// Register отправляет запрос на регистрацию нового пользователя.
//
// Параметры:
//   - ctx: контекст для управления таймаутами и отменой.
//   - login: уникальный логин (например, email).
//   - password: пароль (будет зашифрован на сервере).
//
// Возвращает:
//   - user_id: идентификатор нового пользователя.
//   - ошибка, если:
//   - логин уже существует,
//   - соединение потеряно,
//   - сервер вернул ошибку.
//
// Пример:
//
//	userID, err := client.Register(ctx, "bob@example.com", "mypass")
//	if err != nil {
//	    log.Fatal("Регистрация не удалась:", err)
//	}
func (c *AuthServiceClient) Register(ctx context.Context, login, password string) (string, error) {
	resp, err := c.client.Register(ctx, &pb.RegisterRequest{
		Login:    login,
		Password: password,
	})
	if err != nil {
		return "", err
	}
	return resp.UserId, nil
}

// Login отправляет запрос на аутентификацию.
//
// Проверяет логин и пароль. При успехе возвращает JWT-токен.
//
// Параметры:
//   - ctx: контекст.
//   - login: логин пользователя.
//   - password: пароль.
//
// Возвращает:
//   - access_token: JWT-токен для доступа к защищённым методам.
//   - ошибка, если:
//   - логин/пароль неверны,
//   - пользователь не найден,
//   - сервер недоступен.
//
// Пример:
//
//	token, err := client.Login(ctx, "bob@example.com", "mypass")
//	if err != nil {
//	    log.Fatal("Вход не удался:", err)
//	}
//	fmt.Println("Токен:", token)
func (c *AuthServiceClient) Login(ctx context.Context, login, password string) (string, error) {
	resp, err := c.client.Login(ctx, &pb.LoginRequest{
		Login:    login,
		Password: password})
	if err != nil {
		return "", err
	}
	return resp.AccessToken, nil
}
