package model

import "time"

// User — пользователь системы.
type User struct {
	Login    string `json:"login" db:"login"`
	Password string `json:"password" db:"password"`
}

// AuthData — данные для входа на сайт (логин/пароль).
type AuthData struct {
	ID        string
	UserID    string
	Login     string `json:"login" db:"login"`
	Password  []byte `json:"password" db:"password"`
	Meta      []byte `json:"meta" db:"meta"`
	Version   int64
	UpdatedAt time.Time
}

// FileMeta — файла (файлы хранятся по частям).
type FileMeta struct {
	FileID         string
	CurrentVersion int64
	Filename       string
	SizeBytes      int64
	Checksum       string
	MetaJSON       []byte
}

// TextData — произвольный текст (например, заметки).
type TextData struct {
	ID        string
	UserID    string
	Data      string `json:"data" db:"data"`
	Meta      []byte `json:"meta" db:"meta"`
	Version   int64
	UpdatedAt time.Time
}

// BankCardData — данные банковской карты.
type BankCardData struct {
	ID              string
	UserID          string
	Last4           uint32 `json:"last4" db:"last4"` // последние 4 цифры
	NumberEncrypted []byte `json:"-" db:"number_card"`
	ExpMonth        uint32 `json:"exp_month" db:"exp_month"`
	ExpYear         uint32 `json:"exp_year" db:"exp_year"`
	Owner           string `json:"owner" db:"owner"`
	Meta            []byte `json:"meta" db:"meta"`
	Version         int64
	UpdatedAt       time.Time
}
