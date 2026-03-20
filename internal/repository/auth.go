package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/konkovaanna23/gophkeeper/internal/model"
)

var insertAuthData string = `
       	INSERT INTO storage.auth_data(
			id, user_id,login,password, meta
		)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING version`

var selectAuthData string = `
	 SELECT login, password, meta, version, updated_at
     FROM storage.auth_data 
     WHERE id = $1
		  AND user_id = $2
		  AND deleted_at IS NULL
`

var updateAuthData string = `
	 UPDATE  storage.auth_data 
		SET 
		    login = $1,
		    password = $2,
		    meta = $3,
		    version = version + 1,
		    updated_at = now()
		WHERE id = $4
		  AND user_id = $5
		  AND version = $6
		  AND deleted_at IS NULL
		RETURNING version
`

var deleteAuthData string = `
	 UPDATE storage.auth_data 
		SET deleted_at = now(),
		    version = version + 1,
		    updated_at = now()
		WHERE id = $1
		  AND user_id = $2
		  AND version = $3
		  AND deleted_at IS NULL
 `

func (ds *DBStore) CreateAuthData(ctx context.Context, auth *model.AuthData) (int64, error) {
	var version int64

	err := ds.database.QueryRowContext(ctx, insertAuthData,
		auth.ID, auth.UserID, auth.Login, auth.Password, auth.Meta).
		Scan(&version)
	if err != nil {
		return 0, fmt.Errorf("failed to insert auth data: %w", err)
	}
	return version, nil
}

func (ds *DBStore) GetAuthData(ctx context.Context, userID, authID string) (*model.AuthData, error) {
	auth := &model.AuthData{}
	auth.ID = authID
	auth.UserID = userID

	err := ds.database.QueryRowContext(ctx, selectAuthData, authID, userID).Scan(
		&auth.Login,
		&auth.Password,
		&auth.Meta,
		&auth.Version,
		&auth.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrorNotContent
	}
	return auth, err
}

func (ds *DBStore) UpdateAuthData(ctx context.Context, auth *model.AuthData, expectedVersion int64) (int64, error) {
	var newVersion int64

	err := ds.database.QueryRowContext(ctx, updateAuthData,
		auth.Login, auth.Password, auth.Meta,
		auth.ID, auth.UserID, expectedVersion,
	).Scan(&newVersion)

	if err == sql.ErrNoRows {
		return 0, ErrorVersionConflict
	}
	if err != nil {
		return 0, fmt.Errorf("failed to update auth data: %w", err)
	}

	return newVersion, nil
}

func (ds *DBStore) DeleteAuthData(ctx context.Context, userID, authID string, expectedVersion int64) error {
	res, err := ds.database.ExecContext(ctx, deleteAuthData,
		authID, userID, expectedVersion,
	)
	if err != nil {
		return fmt.Errorf("failed to delete auth data: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to delete auth data: %w", err)
	}
	if rows == 0 {
		return ErrorVersionConflict
	}

	return nil
}
