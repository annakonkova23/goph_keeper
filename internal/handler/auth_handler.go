package handler

import (
	"encoding/json"
	"net/http"

	"github.com/konkovaanna23/gophkeeper/internal/handler/helper"
	"github.com/konkovaanna23/gophkeeper/internal/handler/middleware"
)

type RegisterRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

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

	w.WriteHeader(http.StatusOK)
}
