package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type CRUDRepository[T any] interface {
	Create(ctx context.Context, v *T) (int64, error)
	Get(ctx context.Context, userID, id string) (*T, error)
	Update(ctx context.Context, v *T, expectedVersion int64) (int64, error)
	Delete(ctx context.Context, userID, id string, expectedVersion int64) error
}

type entitySpec[T any] struct {
	insertSQL string
	selectSQL string
	updateSQL string
	deleteSQL string

	insertArgs func(*T) []any
	updateArgs func(*T, int64) []any
	scanDest   func(*T) []any
}

type sqlCRUDRepo[T any] struct {
	db   *sqlx.DB
	spec entitySpec[T]
}

func (r sqlCRUDRepo[T]) Create(ctx context.Context, v *T) (int64, error) {
	var version int64

	err := r.db.QueryRowContext(ctx, r.spec.insertSQL, r.spec.insertArgs(v)...).Scan(&version)
	if err != nil {
		return 0, fmt.Errorf("ошибка добавления в БД: %w", err)
	}
	return version, nil
}

func (r sqlCRUDRepo[T]) Get(ctx context.Context, userID, id string) (*T, error) {
	v := new(T)

	args := []any{id, userID}
	err := r.db.QueryRowContext(ctx, r.spec.selectSQL, args...).Scan(r.spec.scanDest(v)...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrorNotContent
		}
		return nil, fmt.Errorf("ошибка получения данных из БД: %w", err)
	}
	return v, nil
}

func (r sqlCRUDRepo[T]) Update(ctx context.Context, v *T, expectedVersion int64) (int64, error) {
	var newVersion int64

	err := r.db.QueryRowContext(ctx, r.spec.updateSQL, r.spec.updateArgs(v, expectedVersion)...).Scan(&newVersion)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrorVersionConflict
		}
		return 0, fmt.Errorf("ошибка обновления в БД: %w", err)
	}
	return newVersion, nil
}

func (r sqlCRUDRepo[T]) Delete(ctx context.Context, userID, id string, expectedVersion int64) error {
	res, err := r.db.ExecContext(ctx, r.spec.deleteSQL, id, userID, expectedVersion)
	if err != nil {
		return fmt.Errorf("ошибка удаления из БД: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("ошибка удаления из БД: %w", err)
	}
	if rows == 0 {
		return ErrorVersionConflict
	}
	return nil
}
