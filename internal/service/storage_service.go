package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/konkovaanna23/gophkeeper/internal/auth"
	"github.com/konkovaanna23/gophkeeper/internal/repository"
	pb "github.com/konkovaanna23/gophkeeper/pkg/keeperservice"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type StorageServer struct {
	pb.UnimplementedStorageServiceServer
	repo      repository.StoreRepository
	key       []byte
	lgr       *zap.Logger
	tmDelFile time.Duration
}

func NewStorageServer(logger *zap.Logger, repo repository.StoreRepository, key []byte, timeDeleteFile int) *StorageServer {
	return &StorageServer{
		repo:      repo,
		key:       key,
		lgr:       logger,
		tmDelFile: time.Duration(timeDeleteFile) * time.Minute,
	}
}

func (s *StorageServer) StartOfServiceProcesses(ctx context.Context) {

	go s.deleteFrozenFileDownloads(ctx)
}

func (s *StorageServer) deleteFrozenFileDownloads(ctx context.Context) {
	s.lgr.Info("Старт очистки зависших загрузок файла")

	ticker := time.NewTicker(s.tmDelFile)

	for {
		select {
		case <-ctx.Done():
			s.lgr.Info("Отмена контекста. Прерываем очистку зависших загрузок файла")

		case <-ticker.C:
			err := s.repo.Files().CleanupExpiredUploads(ctx)
			if err != nil {
				s.lgr.Error("Ошибка очистки зависших загрузок файла", zap.Error(err))
			}
		}
	}

}

func grpcErr(err error) error {
	switch {
	case errors.Is(err, repository.ErrorNotContent):
		return status.Error(codes.NotFound, "строки не найдены")
	case errors.Is(err, repository.ErrorVersionConflict):
		return status.Error(codes.Aborted, "конфликт версий")
	case errors.Is(err, repository.ErrorEmptyUpload):
		return status.Error(codes.FailedPrecondition, "обновление файла не имеет больше кусков")
	default:
		return status.Error(codes.Internal, err.Error())
	}
}

func userIDFromCtx(ctx context.Context) (string, error) {
	userID, ok := ctx.Value(auth.UserIDKey).(string)
	if !ok || userID == "" {
		return "", status.Error(codes.Unauthenticated, "пользователь не авторизован")
	}
	return userID, nil
}

func marshalMeta(meta map[string]string) ([]byte, error) {
	if meta == nil {
		return []byte("{}"), nil
	}

	b, err := json.Marshal(meta)
	if err != nil {
		return nil, fmt.Errorf("ошибка marshal meta: %w", err)
	}

	return b, nil
}

func unmarshalMeta(raw []byte) (map[string]string, error) {
	if len(raw) == 0 {
		return map[string]string{}, nil
	}

	var meta map[string]string
	if err := json.Unmarshal(raw, &meta); err != nil {
		return nil, fmt.Errorf("ошибка unmarshal meta: %w", err)
	}

	if meta == nil {
		meta = map[string]string{}
	}

	return meta, nil
}

func parseLast4(number string) (uint32, error) {
	if len(number) < 4 {
		return 0, status.Error(codes.InvalidArgument, "номер карты должен содержать минимум 4 цифры")
	}

	last4 := number[len(number)-4:]
	n, err := strconv.ParseUint(last4, 10, 32)
	if err != nil {
		return 0, status.Error(codes.InvalidArgument, "номер карты должен заканчиваться цифрами")
	}

	return uint32(n), nil
}
