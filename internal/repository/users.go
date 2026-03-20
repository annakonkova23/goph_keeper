package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
)

var insertUser string = `
		INSERT INTO storage.users (id, login, password)
		VALUES ($1, $2, $3);
	`
var selectUserLogin string = `SELECT id, password FROM storage.users WHERE login = $1`

func (ds *DBStore) CreateUser(ctx context.Context, id, login, password string) error {

	_, err := ds.database.ExecContext(ctx, insertUser, login, password)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok {
			if pgErr.Code == "23505" {
				return ErrorConflict
			}
		}
		return fmt.Errorf("failed to insert user: %w", err)
	}

	return nil
}

func (ds *DBStore) GetUserByLogin(ctx context.Context, loginSrc string) (string, string, error) {
	var id, password string

	err := ds.database.QueryRowxContext(ctx, selectUserLogin, loginSrc).Scan(&id, &password)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", "", ErrorUserNotFound
		}
		return "", "", fmt.Errorf("database error: %w", err)
	}

	return id, password, nil
}
