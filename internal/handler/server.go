// Package handler - модуль для работы с HTTP-запросами.
package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/konkovaanna23/gophkeeper/internal/client"
	"github.com/konkovaanna23/gophkeeper/internal/handler/middleware"
	"github.com/sirupsen/logrus"
	"go.uber.org/zap"
)

const cookieName = "authorization"

// Server - структура сервера.
type Server struct {
	url           string
	mux           *chi.Mux
	authClient    *client.AuthServiceClient
	storageClient *client.StorageServiceClient
	srv           *http.Server
	lgr           *zap.Logger
}

func NewServer(logger *zap.Logger, url string, grpServer string) (*Server, error) {

	mux := chi.NewRouter()
	authClient, err := client.NewAuthServiceClient(grpServer)
	if err != nil {
		logger.Error("ошибка создания клиента для авторизации", zap.Error(err))
		return nil, err
	}

	storageClient, err := client.NewStorageServiceClient(grpServer)
	if err != nil {
		logger.Error("ошибка создания клиента для хранения информации", zap.Error(err))
		return nil, err
	}

	s := &Server{
		mux:           mux,
		url:           url,
		authClient:    authClient,
		storageClient: storageClient,
		lgr:           logger,
	}
	s.mux.Use(middleware.CompressMiddleware)
	s.mux.Use(middleware.LoggingMiddleware(logger))

	s.mux.Post("/api/register", s.Register)
	s.mux.Group(func(pr chi.Router) {
		pr.Use(middleware.WithAuthCookie(cookieName, true, time.Duration(1)*time.Hour))
		pr.Post("/api/login", s.Login)
	})

	s.mux.Group(func(pr chi.Router) {
		pr.Use(middleware.WithGRPCAuthorization(cookieName))
		pr.Post("/api/storage/auth-info", s.CreateAuthInfo)
		pr.Get("/api/storage/auth-info", s.GetAuthInfo)
		pr.Put("/api/storage/auth-info", s.UpdateAuthInfo)
		pr.Delete("/api/storage/auth-info", s.DeleteAuthInfo)

		pr.Post("/api/storage/text-info", s.CreateTextInfo)
		pr.Get("/api/storage/text-info", s.GetTextInfo)
		pr.Put("/api/storage/text-info", s.UpdateTextInfo)
		pr.Delete("/api/storage/text-info", s.DeleteTextInfo)

		pr.Post("/api/storage/bank-card", s.CreateBankCard)
		pr.Get("/api/storage/bank-card", s.GetBankCard)
		pr.Put("/api/storage/bank-card", s.UpdateBankCard)
		pr.Delete("/api/storage/bank-card", s.DeleteBankCard)

		pr.Post("/api/storage/file", s.CreateFile)
		pr.Get("/api/storage/file", s.GetFile)
		pr.Put("/api/storage/file", s.UpdateFile)
		pr.Delete("/api/storage/file", s.DeleteFile)

	})

	s.srv = &http.Server{
		Addr:    url,
		Handler: mux,
	}

	return s, nil
}

// Start запускает HTTP-сервер в отдельной горутине.
func (s *Server) Start(ctx context.Context) error {

	go func() {
		<-ctx.Done()
		if err := s.srv.Shutdown(ctx); err != nil {
			logrus.Error("Ошибка при закрытии сервера:", err)
		}
	}()

	return s.srv.ListenAndServe()
}

func (s *Server) GetHandler() http.Handler {
	return s.mux
}
