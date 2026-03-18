package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v4"
	"github.com/konkovaanna23/gophkeeper/internal/model"
)

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

func (ds *DBStore) SaveTextData(ctx context.Context, user, text model.TextData) error {
	metaJSON, err := json.Marshal(text.Meta)
	if err != nil {
		return fmt.Errorf("failed to marshal meta: %w", err)
	}

	_, err = ds.database.ExecContext(ctx, insertTextData,
		user, text.Title, text.Data, metaJSON)
	if err != nil {
		return fmt.Errorf("failed to insert text data: %w", err)
	}
	return nil
}

func (ds *DBStore) SaveBankCardData(ctx context.Context, user string, bankCard model.BankCardData) error {
	metaJSON, err := json.Marshal(bankCard.Meta)
	if err != nil {
		return fmt.Errorf("failed to marshal meta: %w", err)
	}

	_, err = ds.database.ExecContext(ctx, insertBankCardData,
		user, bankCard.Last4, bankCard.NumberEncrypted, bankCard.ExpMonth, bankCard.ExpYear, bankCard.Owner, metaJSON)
	if err != nil {
		return fmt.Errorf("failed to insert bank card data: %w", err)
	}
	return nil
}

func (ds *DBStore) GetTextData(ctx context.Context, user, title string) ([]*model.TextData, error) {
	rows, err := ds.database.QueryContext(ctx, selectAuthData, user, title)
	if err != nil {
		return nil, fmt.Errorf("failed to query auth data: %w", err)
	}
	defer rows.Close()

	var result []*model.TextData

	for rows.Next() {
		var data model.TextData
		var metaJSON []byte

		err := rows.Scan(&data.Title, &data.Data, &metaJSON)
		if err != nil {
			return nil, fmt.Errorf("failed to scan auth data row: %w", err)
		}

		if len(metaJSON) > 0 {
			if err := json.Unmarshal(metaJSON, &data.Meta); err != nil {
				return nil, fmt.Errorf("failed to unmarshal meta: %w", err)
			}
		} else {
			data.Meta = map[string]string{} // или оставить nil, зависит от модели
		}

		result = append(result, &data)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error during row iteration: %w", err)
	}

	return result, nil
}

func (ds *DBStore) GetBankCardData(ctx context.Context, user string, last4 uint32) ([]*model.BankCardData, error) {
	rows, err := ds.database.QueryContext(ctx, selectAuthData, user, last4)
	if err != nil {
		return nil, fmt.Errorf("failed to query auth data: %w", err)
	}
	defer rows.Close()

	var result []*model.BankCardData

	for rows.Next() {
		var data model.BankCardData
		var metaJSON []byte

		err := rows.Scan(
			&data.Last4, &data.NumberEncrypted,
			&data.ExpMonth, &data.ExpYear, &data.Owner, &metaJSON,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan auth data row: %w", err)
		}

		if len(metaJSON) > 0 {
			if err := json.Unmarshal(metaJSON, &data.Meta); err != nil {
				return nil, fmt.Errorf("failed to unmarshal meta: %w", err)
			}
		} else {
			data.Meta = map[string]string{}
		}

		result = append(result, &data)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error during row iteration: %w", err)
	}

	return result, nil
}
