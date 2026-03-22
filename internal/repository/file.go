package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"iter"

	"github.com/jmoiron/sqlx"
	"github.com/konkovaanna23/gophkeeper/internal/model"
)

type fileStore struct {
	database *sqlx.DB
}

func newFilesRepo(db *sqlx.DB) FileStore {
	return &fileStore{database: db}
}

var insertFiles string = `
	INSERT INTO storage.files(id, user_id, current_version)
	VALUES ($1, $2, 0)
`

var checkFiles string = `
			SELECT 1
			FROM storage.files
			WHERE id = $1
			  AND user_id = $2
			  AND deleted_at IS NULL
		`

var insertUploads string = `
		INSERT INTO storage.uploads(id, file_id, user_id, expected_version, filename, meta, status)
		VALUES ($1, $2, $3, $4, $5, $6, 'pending')
	`
var insertUploadsChunks string = `
		INSERT INTO storage.upload_chunks(upload_id, chunk_no, data, size_bytes)
		SELECT u.id, $2, $3::BYTEA, octet_length($3::BYTEA)
		FROM storage.uploads u
		WHERE u.id = $1
		  AND u.user_id = $4
		  AND u.status = 'pending'
		ON CONFLICT (upload_id, chunk_no)
		DO UPDATE SET
			data = EXCLUDED.data,
			size_bytes = EXCLUDED.size_bytes
	`

var selectUpload string = `
		SELECT file_id, expected_version, filename, meta
		FROM storage.uploads
		WHERE id = $1
		  AND user_id = $2
		  AND status = 'pending'
		FOR UPDATE
	`

var selectCurrentVersion string = `
		SELECT current_version
		FROM storage.files
		WHERE id = $1
		  AND user_id = $2
		  AND deleted_at IS NULL
		FOR UPDATE
	`

var selectCountChunks string = `
		SELECT COALESCE(SUM(size_bytes), 0), COUNT(*)
		FROM storage.upload_chunks
		WHERE upload_id = $1
	`

var insertFileVersions string = `
		INSERT INTO storage.file_versions(file_id, version, filename, size_bytes, checksum, meta)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
var insertFileChunks string = `
		INSERT INTO storage.file_chunks(file_id, version, chunk_no, data, size_bytes)
		SELECT $1, $2, uc.chunk_no, uc.data, uc.size_bytes
		FROM storage.upload_chunks uc
		WHERE uc.upload_id = $3
	`

var updateVersionFiles string = `
		UPDATE storage.files
		SET current_version = $2
		WHERE id = $1
	`

var updateStatusUploads string = `
		UPDATE storage.uploads
		SET status = 'committed'
		WHERE id = $1
	`

var selectFiles string = `
		SELECT f.id, f.current_version, fv.filename, fv.size_bytes, fv.checksum, fv.meta
		FROM storage.files f
		JOIN storage.file_versions fv
		  ON fv.file_id = f.id
		 AND fv.version = f.current_version
		WHERE f.id = $1
		  AND f.user_id = $2
		  AND f.deleted_at IS NULL
	`

var selectExistsVersion string = `
		SELECT 1
		FROM storage.file_versions fv
		JOIN storage.files f ON f.id = fv.file_id
		WHERE fv.file_id = $1
		  AND fv.version = $2
		  AND f.user_id = $3
		  AND f.deleted_at IS NULL
`

var selectChunkData string = `
		SELECT chunk_no, data
		FROM storage.file_chunks
		WHERE file_id = $1
		  AND version = $2
		ORDER BY chunk_no
	`

var updateDeleteFiles string = `
		UPDATE storage.files
		SET deleted_at = now()
		WHERE id = $1
		  AND user_id = $2
		  AND current_version = $3
		  AND deleted_at IS NULL
	`

func (ds *fileStore) StartUpload(ctx context.Context, userID string, fileID string, expectedVersion int64, filename string, meta map[string]string, newID func() string) (string, string, error) {
	var uploadID, outFileID string
	metaJSON, err := json.Marshal(meta)
	if err != nil {
		return "", "", err
	}

	tx, err := ds.database.BeginTx(ctx, nil)
	if err != nil {
		return "", "", err
	}
	defer tx.Rollback()

	outFileID = fileID
	if outFileID == "" {
		outFileID = newID()
		expectedVersion = 0

		_, err = tx.ExecContext(ctx, insertFiles, outFileID, userID)
		if err != nil {
			return "", "", err
		}
	} else {
		var dummy int
		err = tx.QueryRowContext(ctx, checkFiles, outFileID, userID).Scan(&dummy)
		if errors.Is(err, sql.ErrNoRows) {
			return "", "", ErrorNotContent
		}
		if err != nil {
			return "", "", err
		}
	}

	uploadID = newID()

	_, err = tx.ExecContext(ctx, insertUploads, uploadID, outFileID, userID, expectedVersion, filename, metaJSON)
	if err != nil {
		return "", "", err
	}

	if err := tx.Commit(); err != nil {
		return "", "", err
	}

	return uploadID, outFileID, nil
}

func (ds *fileStore) PutUploadChunk(ctx context.Context, userID string, uploadID string, chunkNo int64, data []byte) error {
	res, err := ds.database.ExecContext(ctx, insertUploadsChunks, uploadID, chunkNo, data, userID)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrorNotContent
	}

	return nil
}

func (ds *fileStore) CommitUpload(ctx context.Context, userID string, uploadID string, fileID string, expectedVersion int64, checksum string) (string, int64, error) {
	tx, err := ds.database.BeginTx(ctx, nil)
	if err != nil {
		return "", 0, err
	}
	defer tx.Rollback()

	var sessFileID string
	var sessExpectedVersion int64
	var filename string
	var metaJSON []byte

	err = tx.QueryRowContext(ctx, selectUpload, uploadID, userID).Scan(&sessFileID, &sessExpectedVersion, &filename, &metaJSON)
	if errors.Is(err, sql.ErrNoRows) {
		return "", 0, ErrorNotContent
	}
	if err != nil {
		return "", 0, err
	}

	if fileID != "" && fileID != sessFileID {
		return "", 0, ErrorVersionConflict
	}
	if expectedVersion != sessExpectedVersion {
		return "", 0, ErrorVersionConflict
	}

	var currentVersion int64
	err = tx.QueryRowContext(ctx, selectCurrentVersion, sessFileID, userID).Scan(&currentVersion)
	if errors.Is(err, sql.ErrNoRows) {
		return "", 0, ErrorNotContent
	}
	if err != nil {
		return "", 0, err
	}

	if currentVersion != sessExpectedVersion {
		return "", 0, ErrorVersionConflict
	}

	var totalSize int64
	var chunksCount int64
	err = tx.QueryRowContext(ctx, selectCountChunks, uploadID).Scan(&totalSize, &chunksCount)
	if err != nil {
		return "", 0, err
	}
	if chunksCount == 0 {
		return "", 0, ErrorEmptyUpload
	}

	newVersion := currentVersion + 1

	_, err = tx.ExecContext(ctx, insertFileVersions, sessFileID, newVersion, filename, totalSize, checksum, metaJSON)
	if err != nil {
		return "", 0, err
	}

	_, err = tx.ExecContext(ctx, insertFileChunks, sessFileID, newVersion, uploadID)
	if err != nil {
		return "", 0, err
	}

	_, err = tx.ExecContext(ctx, updateVersionFiles, sessFileID, newVersion)
	if err != nil {
		return "", 0, err
	}

	_, err = tx.ExecContext(ctx, updateStatusUploads, uploadID)
	if err != nil {
		return "", 0, err
	}

	if err := tx.Commit(); err != nil {
		return "", 0, err
	}

	return sessFileID, newVersion, nil
}

func (ds *fileStore) GetFileMeta(ctx context.Context, userID string, fileID string) (*model.FileMeta, error) {
	var meta model.FileMeta

	err := ds.database.QueryRowContext(ctx, selectFiles, fileID, userID).Scan(
		&meta.FileID,
		&meta.CurrentVersion,
		&meta.Filename,
		&meta.SizeBytes,
		&meta.Checksum,
		&meta.MetaJSON,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	if err != nil {
		return nil, err
	}

	return &meta, nil
}

func (ds *fileStore) resolveVersion(ctx context.Context, userID string, fileID string, version int64) (int64, error) {
	if version == 0 {
		var currentVersion int64
		err := ds.database.QueryRowContext(ctx, selectCurrentVersion, fileID, userID).Scan(&currentVersion)
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrorNotContent
		}
		if err != nil {
			return 0, err
		}
		if currentVersion == 0 {
			return 0, ErrorNotContent
		}
		return currentVersion, nil
	}

	var dummy int
	err := ds.database.QueryRowContext(ctx, selectExistsVersion, fileID, version, userID).Scan(&dummy)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, ErrorNotContent
	}
	if err != nil {
		return 0, err
	}

	return version, nil
}

func (ds *fileStore) StreamFileChunks(ctx context.Context, userID string, fileID string, version int64, fn func(chunkNo int64, data []byte) error) error {
	resolvedVersion, err := ds.resolveVersion(ctx, userID, fileID, version)
	if err != nil {
		return err
	}

	rows, err := ds.database.QueryContext(ctx, selectChunkData, fileID, resolvedVersion)
	if err != nil {
		return err
	}
	defer rows.Close()

	found := false
	for rows.Next() {
		found = true

		var chunkNo int64
		var data []byte

		if err := rows.Scan(&chunkNo, &data); err != nil {
			return err
		}
		if err := fn(chunkNo, data); err != nil {
			return err
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}
	if !found {
		return ErrorNotContent
	}

	return nil
}

func (ds *fileStore) DeleteFile(ctx context.Context, userID string, fileID string, expectedVersion int64) error {
	res, err := ds.database.ExecContext(ctx, updateDeleteFiles, fileID, userID, expectedVersion)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrorVersionConflict
	}

	return nil
}

func (ds *fileStore) CleanupExpiredUploads(ctx context.Context) error {
	_, err := ds.database.ExecContext(ctx, `
		DELETE FROM uploads
		WHERE status = 'pending'
		  AND created_at < now() - interval '24 hours'
	`)
	return err
}

func (ds *fileStore) Chunks(ctx context.Context, userID string, fileID string, version int64) iter.Seq2[*FileChunk, error] {
	return func(yield func(*FileChunk, error) bool) {
		err := ds.StreamFileChunks(
			ctx,
			userID,
			fileID,
			version,
			func(chunkNo int64, data []byte) error {
				if !yield(&FileChunk{
					ChunkNo: chunkNo,
					Data:    data,
				}, nil) {
					return context.Canceled
				}
				return nil
			},
		)

		if err != nil && !errors.Is(err, context.Canceled) {
			yield(nil, err)
		}
	}
}
