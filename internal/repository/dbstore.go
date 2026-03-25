// package repository
package repository

import (
	"context"
	"iter"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/konkovaanna23/gophkeeper/internal/model"
	"go.uber.org/zap"
)

// UserStore Репозиторий для работы с аутентификационными данными.
type UserStore interface {
	CreateUser(ctx context.Context, id, login, password string) error
	GetUserByLogin(ctx context.Context, loginSrc string) (string, string, error)
}

// FileStore Репозиторий для работы с файлами.
type FileStore interface {
	StartUpload(ctx context.Context, userID string, fileID string, expectedVersion int64, filename string, meta map[string]string, newID func() string) (string, string, error)
	PutUploadChunk(ctx context.Context, userID string, uploadID string, chunkNo int64, data []byte) error
	CommitUpload(ctx context.Context, userID string, uploadID string, fileID string, expectedVersion int64, checksum string) (string, int64, error)
	GetFileMeta(ctx context.Context, userID string, fileID string) (*model.FileMeta, error)
	Chunks(ctx context.Context, userID, fileID string, version int64) iter.Seq2[*FileChunk, error]
	DeleteFile(ctx context.Context, userID string, fileID string, expectedVersion int64) error
	CleanupExpiredUploads(ctx context.Context) error
}

// StoreRepository хранилище репозиториев.
type StoreRepository interface {
	Users() UserStore
	Auths() CRUDRepository[model.AuthData]
	Texts() CRUDRepository[model.TextData]
	Cards() CRUDRepository[model.BankCardData]
	Files() FileStore
}

// DBStore хранилище репозиториев.
type DBStore struct {
	database        *sqlx.DB
	timeoutInterval time.Duration
	logger          *zap.Logger

	auths CRUDRepository[model.AuthData]
	texts CRUDRepository[model.TextData]
	cards CRUDRepository[model.BankCardData]
	users UserStore
	files FileStore
}

func (ds *DBStore) Users() UserStore                          { return ds.users }
func (ds *DBStore) Auths() CRUDRepository[model.AuthData]     { return ds.auths }
func (ds *DBStore) Texts() CRUDRepository[model.TextData]     { return ds.texts }
func (ds *DBStore) Cards() CRUDRepository[model.BankCardData] { return ds.cards }
func (ds *DBStore) Files() FileStore                          { return ds.files }

func NewDBStore(logger *zap.Logger, db *sqlx.DB, timeout time.Duration) *DBStore {
	return &DBStore{
		database:        db,
		timeoutInterval: timeout,
		logger:          logger,
		auths:           newAuthRepo(db),
		texts:           newTextRepo(db),
		cards:           newBankCardRepo(db),
		users:           newUsersRepo(db),
		files:           newFilesRepo(db),
	}
}

func newBankCardRepo(db *sqlx.DB) CRUDRepository[model.BankCardData] {
	return sqlCRUDRepo[model.BankCardData]{
		db: db,
		spec: entitySpec[model.BankCardData]{
			insertSQL: insertBankCardData,
			selectSQL: selectBankCardData,
			updateSQL: updateBankCardData,
			deleteSQL: deleteBankCardData,

			insertArgs: func(c *model.BankCardData) []any {
				return []any{
					c.ID, c.UserID, c.Last4, c.NumberEncrypted,
					c.ExpMonth, c.ExpYear, c.Owner, c.Meta,
				}
			},
			updateArgs: func(c *model.BankCardData, expectedVersion int64) []any {
				return []any{
					c.Last4,
					c.NumberEncrypted,
					c.ExpMonth,
					c.ExpYear,
					c.Owner,
					c.Meta,
					c.ID, c.UserID, expectedVersion,
				}
			},
			scanDest: func(c *model.BankCardData) []any {
				return []any{
					&c.Last4,
					&c.NumberEncrypted,
					&c.ExpMonth,
					&c.ExpYear,
					&c.Owner,
					&c.Meta,
					&c.Version,
					&c.UpdatedAt,
				}
			},
		},
	}
}

func newTextRepo(db *sqlx.DB) CRUDRepository[model.TextData] {
	return sqlCRUDRepo[model.TextData]{
		db: db,
		spec: entitySpec[model.TextData]{
			insertSQL: insertTextData,
			selectSQL: selectTextData,
			updateSQL: updateTextData,
			deleteSQL: deleteTextData,

			insertArgs: func(t *model.TextData) []any {
				return []any{t.ID, t.UserID, t.Data, t.Meta}
			},
			updateArgs: func(t *model.TextData, expectedVersion int64) []any {
				return []any{
					t.Data, t.Meta,
					t.ID, t.UserID, expectedVersion,
				}
			},
			scanDest: func(t *model.TextData) []any {
				return []any{
					&t.Data,
					&t.Meta,
					&t.Version,
					&t.UpdatedAt,
				}
			},
		},
	}
}

func newAuthRepo(db *sqlx.DB) CRUDRepository[model.AuthData] {
	return sqlCRUDRepo[model.AuthData]{
		db: db,
		spec: entitySpec[model.AuthData]{
			insertSQL: insertAuthData,
			selectSQL: selectAuthData,
			updateSQL: updateAuthData,
			deleteSQL: deleteAuthData,

			insertArgs: func(a *model.AuthData) []any {
				return []any{a.ID, a.UserID, a.Login, a.Password, a.Meta}
			},
			updateArgs: func(a *model.AuthData, expectedVersion int64) []any {
				return []any{
					a.Login, a.Password, a.Meta,
					a.ID, a.UserID, expectedVersion,
				}
			},
			scanDest: func(a *model.AuthData) []any {
				return []any{
					&a.Login,
					&a.Password,
					&a.Meta,
					&a.Version,
					&a.UpdatedAt,
				}
			},
		},
	}
}
