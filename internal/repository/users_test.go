package repository

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserStore_CreateUser_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	// Оборачиваем *sql.DB в *sqlx.DB
	sqlxDB := sqlx.NewDb(db, "postgres")
	store := newUsersRepo(sqlxDB)

	ctx := context.Background()

	mock.ExpectExec(`INSERT INTO storage\.users`).
		WithArgs("user-123", "alice@example.com", "pass123").
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = store.CreateUser(ctx, "user-123", "alice@example.com", "pass123")

	require.NoError(t, err)
	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestUserStore_CreateUser_Conflict(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	store := newUsersRepo(sqlxDB)

	ctx := context.Background()

	pgErr := &pgconn.PgError{Code: "23505"}
	mock.ExpectExec(`INSERT INTO storage\.users`).
		WithArgs("user-123", "alice@example.com", "pass123").
		WillReturnError(pgErr)

	err = store.CreateUser(ctx, "user-123", "alice@example.com", "pass123")

	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrorConflict))
}

func TestUserStore_CreateUser_OtherDBError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	store := newUsersRepo(sqlxDB)

	ctx := context.Background()

	mock.ExpectExec(`INSERT INTO storage\.users`).
		WithArgs("user-123", "alice@example.com", "pass123").
		WillReturnError(sql.ErrConnDone)

	err = store.CreateUser(ctx, "user-123", "alice@example.com", "pass123")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "ошибка вставки в таблицу storage.users")
	assert.False(t, errors.Is(err, ErrorConflict))
}

func TestUserStore_GetUserByLogin_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	store := newUsersRepo(sqlxDB)

	ctx := context.Background()

	rows := sqlmock.NewRows([]string{"id", "password"}).
		AddRow("user-123", "hashed_pass")

	mock.ExpectQuery(`SELECT id, password FROM storage\.users`).
		WithArgs("alice@example.com").
		WillReturnRows(rows)

	id, password, err := store.GetUserByLogin(ctx, "alice@example.com")

	require.NoError(t, err)
	assert.Equal(t, "user-123", id)
	assert.Equal(t, "hashed_pass", password)
}

func TestUserStore_GetUserByLogin_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	store := newUsersRepo(sqlxDB)

	ctx := context.Background()

	mock.ExpectQuery(`SELECT id, password FROM storage\.users`).
		WithArgs("unknown@example.com").
		WillReturnError(sql.ErrNoRows)

	id, password, err := store.GetUserByLogin(ctx, "unknown@example.com")

	require.Error(t, err)
	assert.Equal(t, "", id)
	assert.Equal(t, "", password)
	assert.True(t, errors.Is(err, ErrorUserNotFound))
}

func TestUserStore_GetUserByLogin_DatabaseError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	store := newUsersRepo(sqlxDB)

	ctx := context.Background()

	mock.ExpectQuery(`SELECT id, password FROM storage\.users`).
		WithArgs("alice@example.com").
		WillReturnError(sql.ErrTxDone)

	id, password, err := store.GetUserByLogin(ctx, "alice@example.com")

	require.Error(t, err)
	assert.Equal(t, "", id)
	assert.Equal(t, "", password)
	assert.Contains(t, err.Error(), "ошибка получения пользователя из БД")
}
