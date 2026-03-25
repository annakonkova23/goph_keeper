package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"
)

type userStore struct {
	database *sqlx.DB
}

func newUsersRepo(db *sqlx.DB) UserStore {
	return &userStore{database: db}
}

var insertUser string = `
		INSERT INTO storage.users (id, login, password)
		VALUES ($1, $2, $3);
	`
var selectUserLogin string = `SELECT id, password FROM storage.users WHERE login = $1`

// CreateUser - добавление в storage.users.
func (ds *userStore) CreateUser(ctx context.Context, id, login, password string) error {

	_, err := ds.database.ExecContext(ctx, insertUser, id, login, password)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok {
			if pgErr.Code == "23505" {
				return ErrorConflict
			}
		}
		return fmt.Errorf("ошибка вставки в таблицу storage.users: %w", err)
	}

	return nil
}

// GetUserByLogin - получение пользователя по логину.
func (ds *userStore) GetUserByLogin(ctx context.Context, loginSrc string) (string, string, error) {
	var id, password string

	err := ds.database.QueryRowxContext(ctx, selectUserLogin, loginSrc).Scan(&id, &password)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", "", ErrorUserNotFound
		}
		return "", "", fmt.Errorf("ошибка получения пользователя из БД: %w", err)
	}

	return id, password, nil
}
