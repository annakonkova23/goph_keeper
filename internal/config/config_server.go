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
	defaultGrpc = "localhost:8080"
)

// Config Конфигурация приложения.
type ConfigServer struct {
	DSN        string `json:"dsn"`
	GrpcServer string `json:"grpc_server_address"`
}

type configServerPointer struct {
	DSN        *string `json:"dsn"`
	GrpcServer *string `json:"grpc_server_address"`
	ConfigPath *string
}

func defaultConfig() *ConfigServer {
	return &ConfigServer{
		DSN:        "", //"postgres://user_main:user_main@localhost:5432/shortenerdb?sslmode=disable",
		GrpcServer: defaultGrpc,
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

	cfg := &ConfigServer{}

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
	configJSONFlag := flag.String("c", "", "Файл конфигурации")

	flag.Parse()

	return &configServerPointer{
		DSN:        dsnFlag,
		ConfigPath: configJSONFlag,
		GrpcServer: grpcServerFlag,
	}
}

func applyEnv(cfg *ConfigServer) {

	if v, ok := lookupEnvString("DSN"); ok {
		cfg.DSN = v
	}

	if v, ok := lookupEnvString("GRPC_SERVER_ADDRESS"); ok {
		cfg.GrpcServer = v
	}
}

func applyFlag(dst *ConfigServer, src *configServerPointer) {
	flag.CommandLine.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "d":
			dst.DSN = *src.DSN
		case "g":
			dst.GrpcServer = *src.GrpcServer
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
}
