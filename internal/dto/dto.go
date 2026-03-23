//package dto пакет для запросов ответов для swagger
package dto

// RegisterRequest — запрос на регистрацию
type RegisterRequest struct {
	Login    string `json:"login" example:"alice@example.com"`
	Password string `json:"password" example:"mysecretpass"`
}

// LoginRequest — запрос на вход
type LoginRequest struct {
	Login    string `json:"login" example:"alice@example.com"`
	Password string `json:"password" example:"mysecretpass"`
}

// LoginResponse — успешный ответ при входе
type LoginResponse struct {
	AccessToken string `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}

// CreateResponse — ответ при создании записи
type CreateResponse struct {
	ID      string `json:"id" example:"rec-123"`
	Version int64  `json:"version" example:"1"`
}

// UpdateResponse — ответ при обновлении
type UpdateResponse struct {
	NewVersion int64 `json:"new_version" example:"2"`
}

// ErrorResponse — ответ при ошибке
type ErrorResponse struct {
	Error string `json:"error" example:"пользователь не найден"`
}
