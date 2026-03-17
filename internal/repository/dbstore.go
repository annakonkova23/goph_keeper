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
	SaveFileChunk(ctx context.Context, user string, file *model.FileChunk) error
	SaveTextData(ctx context.Context, user string, text *model.TextData) error
	SaveBankCardData(ctx context.Context, user string, bankCard *model.BankCardData) error
	GetFileChunk(ctx context.Context, user string, fileName string, chunkNum int) (*model.FileChunk, error)
	GetTextData(ctx context.Context, user string, title string) ([]*model.TextData, error)
	GetBankCardData(ctx context.Context, user string, encryptedNumber uint32) ([]*model.BankCardData, error)
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
