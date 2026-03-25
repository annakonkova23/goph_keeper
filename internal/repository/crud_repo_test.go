package repository

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/konkovaanna23/gophkeeper/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Пример данных
var (
	testNow = time.Now().UTC()
)

func TestSQLCRUDRepo_Create_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := newAuthRepo(sqlxDB)

	ctx := context.Background()
	auth := &model.AuthData{
		ID:       "auth-1",
		UserID:   "user-1",
		Login:    "alice@example.com",
		Password: []byte("encrypted-pass"),
		Meta:     []byte(`{"cat":"work"}`),
	}

	mock.ExpectQuery(regexp.QuoteMeta(insertAuthData)).
		WithArgs(auth.ID, auth.UserID, auth.Login, auth.Password, auth.Meta).
		WillReturnRows(sqlmock.NewRows([]string{"version"}).AddRow(int64(1)))

	version, err := repo.Create(ctx, auth)

	require.NoError(t, err)
	assert.Equal(t, int64(1), version)

	// Проверяем, что все ожидания выполнены
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSQLCRUDRepo_Get_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := newAuthRepo(sqlxDB)

	ctx := context.Background()

	rows := sqlmock.NewRows([]string{"login", "password", "meta", "version", "updated_at"}).
		AddRow(
			"alice", []byte("enc-pass"), []byte(`{"cat":"work"}`), int64(1), testNow,
		)

	mock.ExpectQuery(regexp.QuoteMeta(selectAuthData)).
		WithArgs("auth-1", "user-1").
		WillReturnRows(rows)

	auth, err := repo.Get(ctx, "user-1", "auth-1")

	require.NoError(t, err)
	require.NotNil(t, auth)
	assert.Equal(t, "alice", auth.Login)
	assert.Equal(t, []byte("enc-pass"), auth.Password)
	assert.Equal(t, []byte(`{"cat":"work"}`), auth.Meta)
	assert.Equal(t, int64(1), auth.Version)
	assert.WithinDuration(t, testNow, auth.UpdatedAt, time.Second)
}

func TestSQLCRUDRepo_Get_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := newAuthRepo(sqlxDB)

	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(selectAuthData)).
		WithArgs("auth-1", "user-1").
		WillReturnError(sql.ErrNoRows)

	auth, err := repo.Get(ctx, "user-1", "auth-1")

	require.Error(t, err)
	assert.Nil(t, auth)
	assert.True(t, errors.Is(err, ErrorNotContent))
}

func TestSQLCRUDRepo_Update_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := newAuthRepo(sqlxDB)

	ctx := context.Background()
	auth := &model.AuthData{
		ID:       "auth-1",
		UserID:   "user-1",
		Login:    "new@example.com",
		Password: []byte("new-enc-pass"),
		Meta:     []byte(`{"cat":"personal"}`),
	}
	expectedVersion := int64(1)

	mock.ExpectQuery(regexp.QuoteMeta(updateAuthData)).
		WithArgs(auth.Login, auth.Password, auth.Meta, auth.ID, auth.UserID, expectedVersion).
		WillReturnRows(sqlmock.NewRows([]string{"version"}).AddRow(int64(2)))

	newVersion, err := repo.Update(ctx, auth, expectedVersion)

	require.NoError(t, err)
	assert.Equal(t, int64(2), newVersion)
}

func TestSQLCRUDRepo_Update_VersionConflict(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := newAuthRepo(sqlxDB)

	ctx := context.Background()
	auth := &model.AuthData{ID: "auth-1", UserID: "user-1"}
	expectedVersion := int64(999)

	mock.ExpectQuery(regexp.QuoteMeta(updateAuthData)).
		WithArgs(auth.Login, auth.Password, auth.Meta, auth.ID, auth.UserID, expectedVersion).
		WillReturnError(sql.ErrNoRows)

	newVersion, err := repo.Update(ctx, auth, expectedVersion)

	require.Error(t, err)
	assert.Equal(t, int64(0), newVersion)
	assert.True(t, errors.Is(err, ErrorVersionConflict))
}

func TestSQLCRUDRepo_Delete_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := newAuthRepo(sqlxDB)

	ctx := context.Background()

	mock.ExpectExec(regexp.QuoteMeta(deleteAuthData)).
		WithArgs("auth-1", "user-1", 1).
		WillReturnResult(sqlmock.NewResult(0, 1)) // 1 строка удалена

	err = repo.Delete(ctx, "user-1", "auth-1", 1)

	require.NoError(t, err)
}

func TestSQLCRUDRepo_Delete_VersionConflict(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := newAuthRepo(sqlxDB)

	ctx := context.Background()

	mock.ExpectExec(regexp.QuoteMeta(deleteAuthData)).
		WithArgs("auth-1", "user-1", 999).
		WillReturnResult(sqlmock.NewResult(0, 0)) // 0 строк удалено

	err = repo.Delete(ctx, "user-1", "auth-1", 999)

	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrorVersionConflict))
}
