// Package config - пакет для работы с конфигурацией.
package config

import (
	"encoding/json"
	"flag"
	"log"
	"os"
	"strconv"

	"github.com/konkovaanna23/gophkeeper/internal/file"
)

const (
	defaultGrpc = ":5051"
)

// Config Конфигурация приложения.
type ConfigServer struct {
	DSN               string `json:"dsn"`
	GrpcServer        string `json:"grpc_server_address"`
	Key               string `json:"crypto_key"`
	DeleteFileTimeout int    `json:"delete_file_timeout"`
	KeyAuth           string `json:"auth_key"`
	TokenTTL          int    `json:"token_TTL"`
}

type configServerPointer struct {
	DSN               *string `json:"dsn"`
	GrpcServer        *string `json:"grpc_server_address"`
	Key               *string `json:"crypto_key"`
	DeleteFileTimeout *int    `json:"delete_file_timeout"`
	KeyAuth           *string `json:"auth_key"`
	TokenTTL          *int    `json:"token_TTL"`
	ConfigPath        *string
}

func defaultConfig() *ConfigServer {
	return &ConfigServer{
		DSN:               "postgres://user_main:user_main@localhost:5432/gophkeeperdb?sslmode=disable",
		GrpcServer:        defaultGrpc,
		Key:               "key",
		KeyAuth:           "key",
		DeleteFileTimeout: 1,
		TokenTTL:          2,
	}
}

func lookupEnvString(key string) (string, bool) {
	v := os.Getenv(key)
	if v == "" {
		return "", false
	}
	return v, true
}

func lookupEnvInt(key string) (int, bool) {
	v := os.Getenv(key)
	if v == "" {
		return 0, false
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		return 0, false
	}
	return i, true
}

func lookupEnvBool(key string) (bool, bool) {
	v := os.Getenv(key)
	if v == "" {
		return false, false
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return false, false
	}
	return b, true
}

// GetConfig возвращает конфигурацию приложения.
// Приоритет: ФЛАГИ > ENV > CONFIG(JSON) > DEFAULTS.
func GetConfig() *ConfigServer {

	cfgFlag := readFlag()

	cfg := defaultConfig()

	configPath := ""
	flag.CommandLine.Visit(func(f *flag.Flag) {
		if f.Name == "c" {
			configPath = *cfgFlag.ConfigPath
		}
	})
	if configPath == "" {
		if v, ok := lookupEnvString("CONFIG"); ok {
			configPath = v
		}
	}

	if configPath != "" {
		fc, err := loadConfigFile(configPath)
		if err != nil {
			log.Println("Ошибка чтения файла конфига:", err.Error())
		} else {
			applyConfigFile(cfg, fc)
		}
	}

	applyEnv(cfg)

	applyFlag(cfg, cfgFlag)

	return cfg
}

func readFlag() *configServerPointer {
	grpcServerFlag := flag.String("g", "", "Адрес gRPC-сервера")
	dsnFlag := flag.String("d", "", "DSN для подключения к базе данных")
	keyFlag := flag.String("k", "", "Ключ для шифрования данных")
	keyAuthFlag := flag.String("a", "", "Ключ для шифрования пользователя")
	deleteTimeoutFlag := flag.Int("df", 0, "Таймаут удаления зависших файлов")
	tokenTTLFlag := flag.Int("t", 0, "Таймаут хранения токена пользователя")
	configJSONFlag := flag.String("c", "", "Файл конфигурации")

	flag.Parse()

	return &configServerPointer{
		DSN:               dsnFlag,
		ConfigPath:        configJSONFlag,
		GrpcServer:        grpcServerFlag,
		Key:               keyFlag,
		KeyAuth:           keyAuthFlag,
		DeleteFileTimeout: deleteTimeoutFlag,
		TokenTTL:          tokenTTLFlag,
	}
}

func applyEnv(cfg *ConfigServer) {

	if v, ok := lookupEnvString("DSN"); ok {
		cfg.DSN = v
	}

	if v, ok := lookupEnvString("GRPC_SERVER_ADDRESS"); ok {
		cfg.GrpcServer = v
	}

	if v, ok := lookupEnvString("CRYPTO_KEY"); ok {
		cfg.Key = v
	}

	if v, ok := lookupEnvString("AUTH_KEY"); ok {
		cfg.KeyAuth = v
	}

	if v, ok := lookupEnvInt("DELETE_TIMEOUT"); ok {
		cfg.DeleteFileTimeout = v
	}

	if v, ok := lookupEnvInt("TOKEN_TTL"); ok {
		cfg.TokenTTL = v
	}
}

func applyFlag(dst *ConfigServer, src *configServerPointer) {
	flag.CommandLine.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "d":
			dst.DSN = *src.DSN
		case "g":
			dst.GrpcServer = *src.GrpcServer
		case "k":
			dst.Key = *src.Key
		case "a":
			dst.KeyAuth = *src.KeyAuth
		case "df":
			dst.DeleteFileTimeout = *src.DeleteFileTimeout
		case "t":
			dst.TokenTTL = *src.TokenTTL
		}
	})
}

func loadConfigFile(jsonFile string) (*configServerPointer, error) {
	data, err := file.ReadFromFile(jsonFile)
	if err != nil {
		return nil, err
	}

	var fc configServerPointer
	if err := json.Unmarshal(data, &fc); err != nil {
		return nil, err
	}
	return &fc, nil
}

func applyConfigFile(dst *ConfigServer, src *configServerPointer) {
	if src.DSN != nil {
		dst.DSN = *src.DSN
	}
	if src.GrpcServer != nil {
		dst.GrpcServer = *src.GrpcServer
	}
	if src.Key != nil {
		dst.Key = *src.Key
	}

	if src.KeyAuth != nil {
		dst.KeyAuth = *src.KeyAuth
	}

	if src.DeleteFileTimeout != nil {
		dst.DeleteFileTimeout = *src.DeleteFileTimeout
	}

	if src.TokenTTL != nil {
		dst.TokenTTL = *src.TokenTTL
	}
}
