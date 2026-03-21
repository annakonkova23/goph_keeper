// Package handler - модуль для работы с HTTP-запросами.
package handler

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/konkovaanna23/gophkeeper/internal/client"
	"github.com/konkovaanna23/gophkeeper/internal/handler/middleware"
	pb "github.com/konkovaanna23/gophkeeper/pkg/keeperservice"
	"github.com/sirupsen/logrus"
	"go.uber.org/zap"
)

// Server - структура сервера.
type Server struct {
	url           string
	mux           *chi.Mux
	authClient    *client.AuthServiceClient
	storageClient *pb.StorageServiceClient
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
	s := &Server{
		mux:        mux,
		url:        url,
		authClient: authClient,
	}
	s.mux.Use(middleware.CompressMiddleware)
	s.mux.Use(middleware.LoggingMiddleware(logger))

	s.mux.Post("/api/register", s.Register)
	s.mux.Post("/api/login", s.Login)
	/*s.mux.Get("/ping", s.ping)
	s.mux.Post("/api/shorten", s.newJSONURL)
	s.mux.Post("/api/shorten/batch", s.newJSONBatchURL)
	s.mux.Get("/api/user/urls", s.getURLForUser)
	s.mux.Delete("/api/user/urls", s.deleteURLForUser)
	s.mux.Get("/api/internal/stats", s.stats)*/

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
