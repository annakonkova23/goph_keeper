package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/konkovaanna23/gophkeeper/internal/model"
)

func (ds *DBStore) StartUpload(ctx context.Context, userID string, fileID string, expectedVersion int64, filename string, meta map[string]string, newID func() string) (string, string, error) {
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

		_, err = tx.ExecContext(ctx, `
			INSERT INTO storage.files(id, user_id, current_version)
			VALUES ($1, $2, 0)
		`, outFileID, userID)
		if err != nil {
			return "", "", err
		}
	} else {
		var dummy int
		err = tx.QueryRowContext(ctx, `
			SELECT 1
			FROM storage.files
			WHERE id = $1
			  AND owner_id = $2
			  AND deleted_at IS NULL
		`, outFileID, userID).Scan(&dummy)
		if errors.Is(err, sql.ErrNoRows) {
			return "", "", ErrorNotContent
		}
		if err != nil {
			return "", "", err
		}
	}

	uploadID = newID()

	_, err = tx.ExecContext(ctx, `
		INSERT INTO uploads(id, file_id, owner_id, expected_version, filename, meta, status)
		VALUES ($1, $2, $3, $4, $5, $6, 'pending')
	`, uploadID, outFileID, userID, expectedVersion, filename, metaJSON)
	if err != nil {
		return "", "", err
	}

	if err := tx.Commit(); err != nil {
		return "", "", err
	}

	return uploadID, outFileID, nil
}

func (ds *DBStore) PutUploadChunk(ctx context.Context, userID string, uploadID string, chunkNo int64, data []byte) error {
	res, err := ds.database.ExecContext(ctx, `
		INSERT INTO upload_chunks(upload_id, chunk_no, data, size_bytes)
		SELECT u.id, $2, $3, octet_length($3)
		FROM uploads u
		WHERE u.id = $1
		  AND u.owner_id = $4
		  AND u.status = 'pending'
		ON CONFLICT (upload_id, chunk_no)
		DO UPDATE SET
			data = EXCLUDED.data,
			size_bytes = EXCLUDED.size_bytes
	`, uploadID, chunkNo, data, userID)
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

func (ds *DBStore) CommitUpload(ctx context.Context, userID string, uploadID string, fileID string, expectedVersion int64, checksum string) (string, int64, error) {
	tx, err := ds.database.BeginTx(ctx, nil)
	if err != nil {
		return "", 0, err
	}
	defer tx.Rollback()

	var sessFileID string
	var sessExpectedVersion int64
	var filename string
	var metaJSON []byte

	err = tx.QueryRowContext(ctx, `
		SELECT file_id, expected_version, filename, meta
		FROM uploads
		WHERE id = $1
		  AND owner_id = $2
		  AND status = 'pending'
		FOR UPDATE
	`, uploadID, userID).Scan(&sessFileID, &sessExpectedVersion, &filename, &metaJSON)
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
	err = tx.QueryRowContext(ctx, `
		SELECT current_version
		FROM files
		WHERE id = $1
		  AND owner_id = $2
		  AND deleted_at IS NULL
		FOR UPDATE
	`, sessFileID, userID).Scan(&currentVersion)
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
	err = tx.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(size_bytes), 0), COUNT(*)
		FROM upload_chunks
		WHERE upload_id = $1
	`, uploadID).Scan(&totalSize, &chunksCount)
	if err != nil {
		return "", 0, err
	}
	if chunksCount == 0 {
		return "", 0, ErrorEmptyUpload
	}

	newVersion := currentVersion + 1

	_, err = tx.ExecContext(ctx, `
		INSERT INTO file_versions(file_id, version, filename, size_bytes, checksum, meta)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, sessFileID, newVersion, filename, totalSize, checksum, metaJSON)
	if err != nil {
		return "", 0, err
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO file_chunks(file_id, version, chunk_no, data, size_bytes)
		SELECT $1, $2, uc.chunk_no, uc.data, uc.size_bytes
		FROM upload_chunks uc
		WHERE uc.upload_id = $3
	`, sessFileID, newVersion, uploadID)
	if err != nil {
		return "", 0, err
	}

	_, err = tx.ExecContext(ctx, `
		UPDATE files
		SET current_version = $2
		WHERE id = $1
	`, sessFileID, newVersion)
	if err != nil {
		return "", 0, err
	}

	_, err = tx.ExecContext(ctx, `
		UPDATE uploads
		SET status = 'committed'
		WHERE id = $1
	`, uploadID)
	if err != nil {
		return "", 0, err
	}

	if err := tx.Commit(); err != nil {
		return "", 0, err
	}

	return sessFileID, newVersion, nil
}

func (ds *DBStore) GetFileMeta(ctx context.Context, userID string, fileID string) (*model.FileMeta, error) {
	var meta model.FileMeta

	err := ds.database.QueryRowContext(ctx, `
		SELECT f.id, f.current_version, fv.filename, fv.size_bytes, fv.checksum, fv.meta
		FROM files f
		JOIN file_versions fv
		  ON fv.file_id = f.id
		 AND fv.version = f.current_version
		WHERE f.id = $1
		  AND f.owner_id = $2
		  AND f.deleted_at IS NULL
	`, fileID, userID).Scan(
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

func (ds *DBStore) resolveVersion(ctx context.Context, userID string, fileID string, version int64) (int64, error) {
	if version == 0 {
		var currentVersion int64
		err := ds.database.QueryRowContext(ctx, `
			SELECT current_version
			FROM files
			WHERE id = $1
			  AND owner_id = $2
			  AND deleted_at IS NULL
		`, fileID, userID).Scan(&currentVersion)
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
	err := ds.database.QueryRowContext(ctx, `
		SELECT 1
		FROM file_versions fv
		JOIN files f ON f.id = fv.file_id
		WHERE fv.file_id = $1
		  AND fv.version = $2
		  AND f.owner_id = $3
		  AND f.deleted_at IS NULL
	`, fileID, version, userID).Scan(&dummy)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, ErrorNotContent
	}
	if err != nil {
		return 0, err
	}

	return version, nil
}

func (ds *DBStore) StreamFileChunks(ctx context.Context, userID string, fileID string, version int64, fn func(chunkNo int64, data []byte) error) error {
	resolvedVersion, err := ds.resolveVersion(ctx, userID, fileID, version)
	if err != nil {
		return err
	}

	rows, err := ds.database.QueryContext(ctx, `
		SELECT chunk_no, data
		FROM file_chunks
		WHERE file_id = $1
		  AND version = $2
		ORDER BY chunk_no
	`, fileID, resolvedVersion)
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

func (ds *DBStore) DeleteFile(ctx context.Context, userID string, fileID string, expectedVersion int64) error {
	res, err := ds.database.ExecContext(ctx, `
		UPDATE files
		SET deleted_at = now()
		WHERE id = $1
		  AND owner_id = $2
		  AND current_version = $3
		  AND deleted_at IS NULL
	`, fileID, userID, expectedVersion)
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

func (ds *DBStore) CleanupExpiredUploads(ctx context.Context) error {
	_, err := ds.database.ExecContext(ctx, `
		DELETE FROM uploads
		WHERE status = 'pending'
		  AND created_at < now() - interval '24 hours'
	`)
	return err
}
