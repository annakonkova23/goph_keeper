package repository

import (
	"context"
	"database/sql"
	"fmt"

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

func (ds *DBStore) CreateTextData(ctx context.Context, text *model.TextData) (int64, error) {
	var version int64

	err := ds.database.QueryRowContext(ctx, insertTextData,
		text.ID, text.UserID, text.Data, text.Meta).
		Scan(&version)
	if err != nil {
		return 0, fmt.Errorf("failed to insert text data: %w", err)
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
		return 0, fmt.Errorf("failed to update auth data: %w", err)
	}

	return newVersion, nil
}

func (ds *DBStore) DeleteTextData(ctx context.Context, userID, textID string, expectedVersion int64) error {
	res, err := ds.database.ExecContext(ctx, deleteTextData,
		textID, userID, expectedVersion,
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

func (ds *DBStore) CreateBankCardData(ctx context.Context, card *model.BankCardData) (int64, error) {
	var version int64

	err := ds.database.QueryRowContext(ctx, insertBankCardData,
		card.ID, card.UserID, card.Last4, card.NumberEncrypted, card.ExpMonth, card.ExpYear, card.Owner, card.Meta).
		Scan(&version)
	if err != nil {
		return 0, fmt.Errorf("failed to insert text data: %w", err)
	}
	return version, nil
}

func (ds *DBStore) GetBankCardData(ctx context.Context, userID, cardID string) (*model.BankCardData, error) {
	card := &model.BankCardData{}
	card.ID = cardID
	card.UserID = userID

	err := ds.database.QueryRowContext(ctx, selectTextData, cardID, userID).Scan(
		&card.Last4,
		&card.NumberEncrypted,
		&card.ExpMonth,
		&card.ExpYear,
		&card.Owner,
		&card.Meta,
		&card.Version,
		&card.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrorNotContent
	}
	return card, err
}

func (ds *DBStore) UpdateBankCardData(ctx context.Context, card *model.BankCardData, expectedVersion int64) (int64, error) {
	var newVersion int64

	err := ds.database.QueryRowContext(ctx, updateBankCardData,
		card.Last4,
		card.NumberEncrypted,
		card.ExpMonth,
		card.ExpYear,
		card.Owner,
		card.Meta,
		card.ID, card.UserID, expectedVersion,
	).Scan(&newVersion)

	if err == sql.ErrNoRows {
		return 0, ErrorVersionConflict
	}
	if err != nil {
		return 0, fmt.Errorf("failed to update auth data: %w", err)
	}

	return newVersion, nil
}

func (ds *DBStore) DeleteBancCardData(ctx context.Context, userID, cardID string, expectedVersion int64) error {
	res, err := ds.database.ExecContext(ctx, deleteBankCardData,
		cardID, userID, expectedVersion,
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
