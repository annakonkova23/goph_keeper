package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/konkovaanna23/gophkeeper/internal/model"
)

var insertBankCardData string = `
       	INSERT INTO storage.bank_card_data(
			id, user_id, last4, number_card, exp_month, exp_year, owner, meta 
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING version`

var selectBankCardData string = `
	 SELECT last4, number_card, exp_month, exp_year, owner, meta, version, updated_at
     FROM storage.bank_card_data
     WHERE id = $1
		  AND user_id = $2
		  AND deleted_at IS NULL
`

var updateBankCardData string = `
	 UPDATE storage.bank_card_data
		SET 
		    last4 = $1, 
            number_card = $2, 
            exp_month = $3, 
            exp_year = $4, 
            owner = $5,
		    meta = $6,
		    version = version + 1,
		    updated_at = now()
		WHERE id = $7
		  AND user_id = $8
		  AND version = $9
		  AND deleted_at IS NULL
		RETURNING version
`

var deleteBankCardData string = `
	 UPDATE storage.bank_card_data 
		SET deleted_at = now(),
		    version = version + 1,
		    updated_at = now()
		WHERE id = $1
		  AND user_id = $2
		  AND version = $3
		  AND deleted_at IS NULL
 `

func (ds *DBStore) CreateBankCardData(ctx context.Context, card *model.BankCardData) (int64, error) {
	var version int64

	err := ds.database.QueryRowContext(ctx, insertBankCardData,
		card.ID, card.UserID, card.Last4, card.NumberEncrypted, card.ExpMonth, card.ExpYear, card.Owner, card.Meta).
		Scan(&version)
	if err != nil {
		return 0, fmt.Errorf("ошибка добавления в БД: %w", err)
	}
	return version, nil
}

func (ds *DBStore) GetBankCardData(ctx context.Context, userID, cardID string) (*model.BankCardData, error) {
	card := &model.BankCardData{}
	card.ID = cardID
	card.UserID = userID

	err := ds.database.QueryRowContext(ctx, selectBankCardData, cardID, userID).Scan(
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
		return 0, fmt.Errorf("ошибка обновления в БД: %w", err)
	}

	return newVersion, nil
}

func (ds *DBStore) DeleteBankCardData(ctx context.Context, userID, cardID string, expectedVersion int64) error {
	res, err := ds.database.ExecContext(ctx, deleteBankCardData,
		cardID, userID, expectedVersion,
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
