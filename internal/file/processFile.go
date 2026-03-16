// Package file содержит функции для работы с файлами.
package file

import (
	"os"
)

// SaveToFile сохраняет данные в файл.
// Параметры:
// - fileName:  имя файла.
// - data: данные для сохранения.
func SaveToFile(fileName string, data []byte) error {
	err := os.WriteFile(fileName, data, 0644)
	if err != nil {
		return err
	}
	return nil
}

// ReadFromFile читает данные из файла.
// Параметры:
// - fileName: имя файла.
// Возвращает:
// - []byte: данные из файла.
// - error: ошибка, если произошла ошибка чтения файла.
func ReadFromFile(fileName string) ([]byte, error) {
	result, err := os.ReadFile(fileName)
	if err != nil {
		return nil, err
	}
	return result, err

}
