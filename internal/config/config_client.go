// Package config - пакет для работы с конфигурацией.
package config

import (
	"encoding/json"
	"flag"
	"log"

	"github.com/konkovaanna23/gophkeeper/internal/file"
)

const (
	defaultHost = "localhost:8080"
)

// Config Конфигурация приложения.
type ConfigClient struct {
	Host       string `json:"host"`
	GrpcServer string `json:"grpc_server_address"`
}

type configClientPointer struct {
	Host       *string `json:"host"`
	GrpcServer *string `json:"grpc_server_address"`
	ConfigPath *string
}

// GetConfig возвращает конфигурацию приложения.
// Приоритет: ФЛАГИ > ENV > CONFIG(JSON) > DEFAULTS.
func GetConfigClient() *ConfigClient {

	cfgFlag := readClientFlag()

	cfg := &ConfigClient{}

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
		fc, err := loadClientConfigFile(configPath)
		if err != nil {
			log.Println("Ошибка чтения файла конфига:", err.Error())
		} else {
			applyClientConfigFile(cfg, fc)
		}
	}

	applyClientEnv(cfg)

	applyClientFlag(cfg, cfgFlag)

	return cfg
}

func readClientFlag() *configClientPointer {
	hostFlag := flag.String("a", defaultHost, "Адрес запуска HTTP-сервера")
	grpcServerFlag := flag.String("g", "", "Адрес gRPC-сервера")
	configJSONFlag := flag.String("c", "", "Файл конфигурации")

	flag.Parse()

	return &configClientPointer{
		Host:       hostFlag,
		GrpcServer: grpcServerFlag,
		ConfigPath: configJSONFlag,
	}
}

func applyClientEnv(cfg *ConfigClient) {

	if v, ok := lookupEnvString("HOST"); ok {
		cfg.Host = v
	}

	if v, ok := lookupEnvString("GRPC_SERVER_ADDRESS"); ok {
		cfg.GrpcServer = v
	}

}

func applyClientFlag(dst *ConfigClient, src *configClientPointer) {
	flag.CommandLine.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "a":
			dst.Host = *src.Host
		case "g":
			dst.GrpcServer = *src.GrpcServer
		}
	})
}

func loadClientConfigFile(jsonFile string) (*configClientPointer, error) {
	data, err := file.ReadFromFile(jsonFile)
	if err != nil {
		return nil, err
	}

	var fc configClientPointer
	if err := json.Unmarshal(data, &fc); err != nil {
		return nil, err
	}
	return &fc, nil
}

func applyClientConfigFile(dst *ConfigClient, src *configClientPointer) {
	if src.Host != nil {
		dst.Host = *src.Host
	}
	if src.GrpcServer != nil {
		dst.GrpcServer = *src.GrpcServer
	}
}
