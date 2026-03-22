package repository

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newMockDB() (*sqlx.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		panic(err)
	}
	return sqlx.NewDb(db, "postgres"), mock
}

func TestFileStore_StartUpload_NewFile(t *testing.T) {
	db, mock := newMockDB()
	defer db.Close()

	repo := newFilesRepo(db)

	ctx := context.Background()
	userID := "user-1"
	fileID := ""
	expectedVersion := int64(0)
	filename := "test.txt"
	meta := map[string]string{"cat": "work"}

	ids := []string{"file-new", "upload-1"}
	idx := 0
	newID := func() string {
		id := ids[idx]
		idx++
		return id
	}

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(insertFiles)).
		WithArgs("file-new", userID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec(regexp.QuoteMeta(insertUploads)).
		WithArgs(
			"upload-1",
			"file-new",
			userID,
			int64(0),
			filename,
			[]byte(`{"cat":"work"}`),
		).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	uploadID, outFileID, err := repo.StartUpload(ctx, userID, fileID, expectedVersion, filename, meta, newID)

	require.NoError(t, err)
	assert.Equal(t, "upload-1", uploadID)
	assert.Equal(t, "file-new", outFileID)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestFileStore_StartUpload_ExistingFile(t *testing.T) {
	db, mock := newMockDB()
	defer db.Close()

	repo := newFilesRepo(db)

	ctx := context.Background()
	userID := "user-1"
	fileID := "file-1"
	expectedVersion := int64(5)
	filename := "update.txt"
	meta := map[string]string{"cat": "work"}

	newID := func() string { return "upload-2" }

	mock.ExpectBegin()

	rows := sqlmock.NewRows([]string{"1"}).AddRow(1)
	mock.ExpectQuery(regexp.QuoteMeta(checkFiles)).
		WithArgs(fileID, userID).
		WillReturnRows(rows)

	mock.ExpectExec(regexp.QuoteMeta(insertUploads)).
		WithArgs("upload-2", fileID, userID, expectedVersion, filename, []byte(`{"cat":"work"}`)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	uploadID, outFileID, err := repo.StartUpload(ctx, userID, fileID, expectedVersion, filename, meta, newID)

	require.NoError(t, err)
	assert.Equal(t, "upload-2", uploadID)
	assert.Equal(t, "file-1", outFileID)
}

func TestFileStore_StartUpload_FileNotFound(t *testing.T) {
	db, mock := newMockDB()
	defer db.Close()

	repo := newFilesRepo(db)

	ctx := context.Background()
	userID := "user-1"
	fileID := "unknown-1"
	newID := func() string { return "upload-3" }

	mock.ExpectQuery(regexp.QuoteMeta(checkFiles)).
		WithArgs(fileID, userID).
		WillReturnError(sql.ErrNoRows)

	_, _, err := repo.StartUpload(ctx, userID, fileID, 0, "test.txt", nil, newID)

	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrorNotContent))
}

func TestFileStore_PutUploadChunk_Success(t *testing.T) {
	db, mock := newMockDB()
	defer db.Close()

	repo := newFilesRepo(db)

	ctx := context.Background()
	uploadID := "upload-1"
	userID := "user-1"
	chunkNo := int64(0)
	data := []byte("chunk-data")

	mock.ExpectExec(regexp.QuoteMeta(insertUploadsChunks)).
		WithArgs(uploadID, chunkNo, data, userID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.PutUploadChunk(ctx, userID, uploadID, chunkNo, data)

	require.NoError(t, err)
	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestFileStore_PutUploadChunk_Fail(t *testing.T) {
	db, mock := newMockDB()
	defer db.Close()

	repo := newFilesRepo(db)

	ctx := context.Background()
	uploadID := "upload-1"
	userID := "user-1"
	chunkNo := int64(0)
	data := []byte("chunk-data")

	mock.ExpectExec(regexp.QuoteMeta(insertUploadsChunks)).
		WithArgs(uploadID, chunkNo, data, userID).
		WillReturnResult(sqlmock.NewResult(0, 0)) // 0 affected

	err := repo.PutUploadChunk(ctx, userID, uploadID, chunkNo, data)

	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrorNotContent))
}

func TestFileStore_CommitUpload_Success(t *testing.T) {
	db, mock := newMockDB()
	defer db.Close()

	repo := newFilesRepo(db)

	ctx := context.Background()
	userID := "user-1"
	uploadID := "upload-1"
	fileID := "file-1"
	expectedVersion := int64(3)
	checksum := "abc123"

	rowsUpload := sqlmock.NewRows([]string{"file_id", "expected_version", "filename", "meta"}).
		AddRow(fileID, expectedVersion, "test.txt", []byte(`{"cat":"work"}`))

	mock.ExpectBegin()

	mock.ExpectQuery(regexp.QuoteMeta(selectUpload)).
		WithArgs(uploadID, userID).
		WillReturnRows(rowsUpload)

	mock.ExpectQuery(regexp.QuoteMeta(selectCurrentVersion)).
		WithArgs(fileID, userID).
		WillReturnRows(sqlmock.NewRows([]string{"current_version"}).AddRow(expectedVersion))

	mock.ExpectQuery(regexp.QuoteMeta(selectCountChunks)).
		WithArgs(uploadID).
		WillReturnRows(sqlmock.NewRows([]string{"total_size", "chunks_count"}).AddRow(int64(12), int64(1)))

	mock.ExpectExec(regexp.QuoteMeta(insertFileVersions)).
		WithArgs(fileID, expectedVersion+1, "test.txt", int64(12), checksum, []byte(`{"cat":"work"}`)).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec(regexp.QuoteMeta(insertFileChunks)).
		WithArgs(fileID, expectedVersion+1, uploadID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec(regexp.QuoteMeta(updateVersionFiles)).
		WithArgs(fileID, expectedVersion+1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec(regexp.QuoteMeta(updateStatusUploads)).
		WithArgs(uploadID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	newFileID, newVersion, err := repo.CommitUpload(ctx, userID, uploadID, fileID, expectedVersion, checksum)

	require.NoError(t, err)
	assert.Equal(t, fileID, newFileID)
	assert.Equal(t, expectedVersion+1, newVersion)
}

func TestFileStore_CommitUpload_VersionConflict(t *testing.T) {
	db, mock := newMockDB()
	defer db.Close()

	repo := newFilesRepo(db)

	ctx := context.Background()
	userID := "user-1"
	uploadID := "upload-1"
	fileID := "file-1"
	expectedVersion := int64(3)

	rowsUpload := sqlmock.NewRows([]string{"file_id", "expected_version", "filename", "meta"}).
		AddRow(fileID, expectedVersion, "test.txt", []byte(`{"cat":"work"}`))
	mock.ExpectQuery(regexp.QuoteMeta(selectUpload)).
		WithArgs(uploadID, userID).
		WillReturnRows(rowsUpload)

	mock.ExpectQuery(regexp.QuoteMeta(selectCurrentVersion)).
		WithArgs(fileID, userID).
		WillReturnRows(sqlmock.NewRows([]string{"current_version"}).AddRow(expectedVersion + 1))

	_, _, err := repo.CommitUpload(ctx, userID, uploadID, fileID, expectedVersion, "abc123")

	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrorVersionConflict))
}

func TestFileStore_GetFileMeta_Success(t *testing.T) {
	db, mock := newMockDB()
	defer db.Close()

	repo := newFilesRepo(db)

	ctx := context.Background()
	userID := "user-1"
	fileID := "file-1"

	rows := sqlmock.NewRows([]string{"id", "current_version", "filename", "size_bytes", "checksum", "meta"}).
		AddRow(fileID, int64(1), "test.txt", int64(12), "abc123", []byte(`{"cat":"work"}`))

	mock.ExpectQuery(regexp.QuoteMeta(selectFiles)).
		WithArgs(fileID, userID).
		WillReturnRows(rows)

	meta, err := repo.GetFileMeta(ctx, userID, fileID)

	require.NoError(t, err)
	require.NotNil(t, meta)
	assert.Equal(t, fileID, meta.FileID)
	assert.Equal(t, int64(1), meta.CurrentVersion)
	assert.Equal(t, "test.txt", meta.Filename)
	assert.Equal(t, []byte(`{"cat":"work"}`), meta.MetaJSON)
}

func TestFileStore_GetFileMeta_NotFound(t *testing.T) {
	db, mock := newMockDB()
	defer db.Close()

	repo := newFilesRepo(db)

	ctx := context.Background()
	userID := "user-1"
	fileID := "unknown"

	mock.ExpectQuery(regexp.QuoteMeta(selectFiles)).
		WithArgs(fileID, userID).
		WillReturnError(sql.ErrNoRows)

	_, err := repo.GetFileMeta(ctx, userID, fileID)

	require.Error(t, err)
	assert.True(t, errors.Is(err, sql.ErrNoRows))
}

func TestFileStore_DeleteFile_Success(t *testing.T) {
	db, mock := newMockDB()
	defer db.Close()

	repo := newFilesRepo(db)

	ctx := context.Background()
	userID := "user-1"
	fileID := "file-1"
	version := int64(5)

	mock.ExpectExec(regexp.QuoteMeta(updateDeleteFiles)).
		WithArgs(fileID, userID, version).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.DeleteFile(ctx, userID, fileID, version)

	require.NoError(t, err)
}

func TestFileStore_DeleteFile_Conflict(t *testing.T) {
	db, mock := newMockDB()
	defer db.Close()

	repo := newFilesRepo(db)

	ctx := context.Background()
	userID := "user-1"
	fileID := "file-1"
	version := int64(5)

	mock.ExpectExec(regexp.QuoteMeta(updateDeleteFiles)).
		WithArgs(fileID, userID, version).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := repo.DeleteFile(ctx, userID, fileID, version)

	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrorVersionConflict))
}
