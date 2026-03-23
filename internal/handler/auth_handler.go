package handler

import (
	"encoding/json"
	"net/http"

	"github.com/konkovaanna23/gophkeeper/internal/handler/helper"
	"github.com/konkovaanna23/gophkeeper/internal/handler/middleware"
)

// RegisterRequest represents registration input
type RegisterRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

// RegisterResponse represents successful creation
type RegisterResponse struct {
	UserID string `json:"user_id"`
}

type LoginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type LoginResponse struct {
	AccessToken string `json:"access_token"`
}

// Register godoc
// @Summary      Регистрация нового пользователя
// @Description  Создаёт нового пользователя в системе
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        input body dto.RegisterRequest true "Данные регистрации"
// @Success      200 {object} dto.CreateResponse
// @Failure      400 {object} dto.ErrorResponse
// @Failure      500 {object} dto.ErrorResponse
// @Router       /register [post]
func (s *Server) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {

		http.Error(w, "Невалидный JSON", http.StatusBadRequest)
		return
	}

	userID, err := s.authClient.Register(r.Context(), req.Login, req.Password)
	if err != nil {
		helper.WriteGRPCError(w, err)
		return
	}

	helper.WriteJSON(w, http.StatusOK, RegisterResponse{UserID: userID})

}

// Login godoc
// @Summary      Аутентификация пользователя
// @Description  Возвращает JWT-токен при успешном входе
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        input body dto.LoginRequest true "Логин и пароль"
// @Success      200 {object} dto.LoginResponse
// @Failure      401 {object} dto.ErrorResponse
// @Failure      500 {object} dto.ErrorResponse
// @Router       /login [post]
func (s *Server) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Невалидный JSON", http.StatusBadRequest)
		return
	}

	token, err := s.authClient.Login(r.Context(), req.Login, req.Password)
	if err != nil {
		helper.WriteGRPCError(w, err)
		return
	}

	middleware.SetAccessToken(r.Context(), token)

	helper.WriteJSON(w, http.StatusOK, LoginResponse{
		AccessToken: token,
	})
}
