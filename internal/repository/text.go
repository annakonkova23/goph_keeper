package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/konkovaanna23/gophkeeper/internal/model"
)

var insertTextData string = `
       	INSERT INTO storage.text_data(
			id, user_id, data, meta
		)
		VALUES ($1, $2, $3, $4)
		RETURNING version`

var selectTextData string = `
	 SELECT data, meta, version, updated_at
     FROM storage.text_data 
     WHERE id = $1
		  AND user_id = $2
		  AND deleted_at IS NULL
`

var updateTextData string = `
	 UPDATE  storage.text_data 
		SET 
		    data = $1,
		    meta = $2,
		    version = version + 1,
		    updated_at = now()
		WHERE id = $3
		  AND user_id = $4
		  AND version = $5
		  AND deleted_at IS NULL
		RETURNING version
`

var deleteTextData string = `
	 UPDATE storage.text_data 
		SET deleted_at = now(),
		    version = version + 1,
		    updated_at = now()
		WHERE id = $1
		  AND user_id = $2
		  AND version = $3
		  AND deleted_at IS NULL
 `

func (ds *DBStore) CreateTextData(ctx context.Context, text *model.TextData) (int64, error) {
	var version int64

	err := ds.database.QueryRowContext(ctx, insertTextData,
		text.ID, text.UserID, text.Data, text.Meta).
		Scan(&version)
	if err != nil {
		return 0, fmt.Errorf("ошибка добавления в БД: %w", err)
	}
	return version, nil
}

func (ds *DBStore) GetTextData(ctx context.Context, userID, textID string) (*model.TextData, error) {
	text := &model.TextData{}
	text.ID = textID
	text.UserID = userID

	err := ds.database.QueryRowContext(ctx, selectTextData, textID, userID).Scan(
		&text.Data,
		&text.Meta,
		&text.Version,
		&text.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrorNotContent
	}
	return text, err
}

func (ds *DBStore) UpdateTextData(ctx context.Context, text *model.TextData, expectedVersion int64) (int64, error) {
	var newVersion int64

	err := ds.database.QueryRowContext(ctx, updateTextData,
		text.Data, text.Meta,
		text.ID, text.UserID, expectedVersion,
	).Scan(&newVersion)

	if err == sql.ErrNoRows {
		return 0, ErrorVersionConflict
	}
	if err != nil {
		return 0, fmt.Errorf("ошибка обновления в БД: %w", err)
	}

	return newVersion, nil
}

func (ds *DBStore) DeleteTextData(ctx context.Context, userID, textID string, expectedVersion int64) error {
	res, err := ds.database.ExecContext(ctx, deleteTextData,
		textID, userID, expectedVersion,
	)
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
