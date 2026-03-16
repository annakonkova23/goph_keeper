package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v4"
	"github.com/konkovaanna23/gophkeeper/internal/model"
)

func (ds *DBStore) SaveAuthData(ctx context.Context, user string, auth model.AuthData) error {
	metaJSON, err := json.Marshal(auth.Meta)
	if err != nil {
		return fmt.Errorf("failed to marshal meta: %w", err)
	}

	_, err = ds.database.ExecContext(ctx, insertAuthData,
		user, auth.Site, auth.Login, auth.Password, metaJSON)
	if err != nil {
		return fmt.Errorf("failed to insert auth data: %w", err)
	}
	return nil
}

func (ds *DBStore) SaveFileChunk(ctx context.Context, user, file model.FileChunk) error {
	metaJSON, err := json.Marshal(file.Meta)
	if err != nil {
		return fmt.Errorf("failed to marshal meta: %w", err)
	}

	_, err = ds.database.ExecContext(ctx, insertFileChunk,
		user, file.FileName, file.ChunkNum, file.Data, metaJSON)
	if err != nil {
		return fmt.Errorf("failed to insert file chunk: %w", err)
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

func (ds *DBStore) GetAuthDataList(ctx context.Context, user, site string) ([]*model.AuthData, error) {
	rows, err := ds.database.QueryContext(ctx, selectAuthData, user, site)
	if err != nil {
		return nil, fmt.Errorf("failed to query auth data: %w", err)
	}
	defer rows.Close()

	var result []*model.AuthData

	for rows.Next() {
		var data model.AuthData
		var metaJSON []byte

		err := rows.Scan(&data.Site, &data.Login, &data.Password, &metaJSON)
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

func (ds *DBStore) GetFileChunk(ctx context.Context, user, fileName string, chunkNum int) (*model.FileChunk, error) {
	var chunk model.FileChunk
	var metaJSON []byte

	err := ds.database.QueryRowContext(ctx, selectFileData, user, fileName, chunkNum).Scan(
		&chunk.FileName, &chunk.ChunkNum, &chunk.Data, &metaJSON,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrorNotContent
		}
		return nil, fmt.Errorf("failed to get file chunk: %w", err)
	}

	if err = json.Unmarshal(metaJSON, &chunk.Meta); err != nil {
		return nil, fmt.Errorf("failed to unmarshal meta: %w", err)
	}

	return &chunk, nil
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
