package model

// User — пользователь системы
type User struct {
	Login    string `json:"login" db:"login"`
	Password string `json:"password" db:"password"`
}

// AuthData — данные для входа на сайт (логин/пароль)
type AuthData struct {
	Site     string            `json:"site" db:"site"`
	Login    string            `json:"login" db:"login"`
	Password string            `json:"password" db:"password"`
	Meta     map[string]string `json:"meta" db:"meta"`
}

// FileChunk — чанк файла (файлы хранятся по частям)
type FileChunk struct {
	FileName string            `json:"file_name" db:"file_name"`
	ChunkNum int               `json:"chunk_num" db:"chunk_num"`
	Data     []byte            `json:"data" db:"data"`
	Meta     map[string]string `json:"meta" db:"meta"`
}

// TextData — произвольный текст (например, заметки)
type TextData struct {
	User  string            `json:"user" db:"user"`
	Title string            `json:"title" db:"title"`
	Data  string            `json:"data" db:"data"`
	Meta  map[string]string `json:"meta" db:"meta"`
}

// BankCardData — данные банковской карты
type BankCardData struct {
	Last4           uint32            `json:"last4" db:"last4"`   // последние 4 цифры
	NumberEncrypted string            `json:"-" db:"number_card"` // зашифрованный номер (не показываем в JSON)
	ExpMonth        uint32            `json:"exp_month" db:"exp_month"`
	ExpYear         uint32            `json:"exp_year" db:"exp_year"`
	Owner           string            `json:"owner" db:"owner"`
	Meta            map[string]string `json:"meta" db:"meta"`
}
