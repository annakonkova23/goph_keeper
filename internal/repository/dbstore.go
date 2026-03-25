package repository

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/konkovaanna23/gophkeeper/internal/model"
	"go.uber.org/zap"
)

type StoreRepository interface {
	CreateUser(ctx context.Context, id, login, password string) error
	GetUserByLogin(ctx context.Context, loginSrc string) (string, string, error)
	CreateAuthData(ctx context.Context, auth *model.AuthData) (int64, error)
	GetAuthData(ctx context.Context, userID, authID string) (*model.AuthData, error)
	UpdateAuthData(ctx context.Context, auth *model.AuthData, expectedVersion int64) (int64, error)
	DeleteAuthData(ctx context.Context, userID, authID string, expectedVersion int64) error
	StartUpload(ctx context.Context, userID string, fileID string, expectedVersion int64, filename string, meta map[string]string, newID func() string) (string, string, error)
	PutUploadChunk(ctx context.Context, userID string, uploadID string, chunkNo int64, data []byte) error
	CommitUpload(ctx context.Context, userID string, uploadID string, fileID string, expectedVersion int64, checksum string) (string, int64, error)
	GetFileMeta(ctx context.Context, userID string, fileID string) (*model.FileMeta, error)
	StreamFileChunks(ctx context.Context, userID string, fileID string, version int64, fn func(chunkNo int64, data []byte) error) error
	DeleteFile(ctx context.Context, userID string, fileID string, expectedVersion int64) error
	CreateTextData(ctx context.Context, text *model.TextData) (int64, error)
	GetTextData(ctx context.Context, userID, textID string) (*model.TextData, error)
	UpdateTextData(ctx context.Context, text *model.TextData, expectedVersion int64) (int64, error)
	DeleteTextData(ctx context.Context, userID, textID string, expectedVersion int64) error
	CreateBankCardData(ctx context.Context, card *model.BankCardData) (int64, error)
	GetBankCardData(ctx context.Context, userID, cardID string) (*model.BankCardData, error)
	UpdateBankCardData(ctx context.Context, card *model.BankCardData, expectedVersion int64) (int64, error)
	DeleteBankCardData(ctx context.Context, userID, cardID string, expectedVersion int64) error
}

type DBStore struct {
	database        *sqlx.DB
	timeoutInterval time.Duration
	logger          *zap.Logger
}

func NewDBStore(logger *zap.Logger, db *sqlx.DB, timeout time.Duration) *DBStore {
	return &DBStore{
		database:        db,
		timeoutInterval: timeout,
		logger:          logger,
	}
}
