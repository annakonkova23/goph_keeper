package config

import (
	"flag"
	"log"
	"os"
	"strconv"
)

const (
	defaultHost            = "localhost:8080"
	defaultStoreInterval   = 300
	defaultRestore         = true
	defaultFileStoragePath = ""
	//defaultFileStoragePath = "metrics.json"
	//defaultDBDSN = ""
	//defaultDBDSN      = "postgres://postgres:anna@localhost:5432/metricsdb?sslmode=disable" //"host=localhost port=5432 user=postgres password=anna dbname=metricsdb sslmode=disable"
	defaultKey        = "key"
	defaultBufferSize = 100
)

type ServerOptions struct {
	Host            string
	StoreInterval   int
	Restore         bool
	FileStoragePath string
	DatabaseDSN     string
	Key             string
	AuditFilePath   string
	AuditURL        string
	BufferSize      int
	KeyPath         string
	FileConfig      string
	TrustedSubnet   string
	GrpcHost        string
}

func getEnvString(envKey, defaultValue string) string {
	if v, exists := os.LookupEnv(envKey); exists {
		return v
	}
	return defaultValue
}

func getEnvInt(envKey string, defaultValue int) int {
	if v, exists := os.LookupEnv(envKey); exists {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return defaultValue
}

func getEnvBool(envKey string, defaultValue bool) bool {
	if v, exists := os.LookupEnv(envKey); exists {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return defaultValue
}

// NewServerOptions создает новый экземпляр ServerOptions с значениями по умолчанию.
func NewServerOptions() *ServerOptions {
	serverOptions := &ServerOptions{}

	serverOptionsFlag := getServerOptionsFromFlag()

	serverOptionsEnv := getServerOptionsFromEnv(serverOptionsFlag)

	serverOptions = serverOptionsEnv

	if serverOptions.FileConfig != "" {
		so, err := getServerOptionsFromFile(serverOptions.FileConfig)
		if err != nil {
			log.Printf("Не удалось считать конфигурацию из файла %s: %s", serverOptions.FileConfig, err.Error())
		}
		serverOptions.CompareAndAddValues(so)
	}

	return serverOptions

}

func getServerOptionsFromEnv(def *ServerOptions) *ServerOptions {

	serverOptions := &ServerOptions{}

	serverOptions.Host = getEnvString("ADDRESS", def.Host)
	serverOptions.StoreInterval = getEnvInt("STORE_INTERVAL", def.StoreInterval)
	serverOptions.Restore = getEnvBool("RESTORE", def.Restore)
	serverOptions.FileStoragePath = getEnvString("FILE_STORAGE_PATH", def.FileStoragePath)
	serverOptions.DatabaseDSN = getEnvString("DATABASE_DSN", def.DatabaseDSN)
	serverOptions.Key = getEnvString("KEY", def.Key)
	serverOptions.AuditFilePath = getEnvString("AUDIT_FILE", def.AuditFilePath)
	serverOptions.AuditURL = getEnvString("AUDIT_URL", def.AuditURL)
	serverOptions.BufferSize = getEnvInt("BUFFER_SIZE", def.BufferSize)
	serverOptions.KeyPath = getEnvString("CRYPTO_KEY", def.KeyPath)
	serverOptions.FileConfig = getEnvString("CONFIG", def.FileConfig)
	serverOptions.TrustedSubnet = getEnvString("TRUSTED_SUBNET", def.TrustedSubnet)
	serverOptions.GrpcHost = getEnvString("GRPC_HOST", def.GrpcHost)
	return serverOptions
}

func getServerOptionsFromFile(path string) (*ServerOptions, error) {
	config, err := ReadJSONConfig(path)
	if err != nil {
		return nil, err
	}

	so := &ServerOptions{}

	for key, value := range config {
		switch key {
		case "address":
			if s, ok := AsString(value); ok {
				so.Host = s
			} else {
				log.Printf("некорректный тип для 'address': %T", value)
			}
		case "store_interval":
			if n, ok := AsInt(value); ok {
				so.StoreInterval = n
			} else {
				log.Printf("некорректный тип для 'store_interval': %T", value)
			}
		case "restore":
			if b, ok := AsBool(value); ok {
				so.Restore = b
			} else {
				log.Printf("некорректный тип для 'restore': %T", value)
			}
		case "file_storage_path":
			if s, ok := AsString(value); ok {
				so.FileStoragePath = s
			} else {
				log.Printf("некорректный тип для 'file_storage_path': %T", value)
			}
		case "database_dsn":
			if s, ok := AsString(value); ok {
				so.DatabaseDSN = s
			} else {
				log.Printf("некорректный тип для 'database_dsn': %T", value)
			}
		case "key":
			if s, ok := AsString(value); ok {
				so.Key = s
			} else {
				log.Printf("некорректный тип для 'key': %T", value)
			}
		case "audit_file":
			if s, ok := AsString(value); ok {
				so.AuditFilePath = s
			} else {
				log.Printf("некорректный тип для 'audit_file': %T", value)
			}
		case "audit_url":
			if s, ok := AsString(value); ok {
				so.AuditURL = s
			} else {
				log.Printf("некорректный тип для 'audit_url': %T", value)
			}
		case "buffer_size":
			if n, ok := AsInt(value); ok {
				so.BufferSize = n
			} else {
				log.Printf("некорректный тип для 'buffer_size': %T", value)
			}
		case "crypto_key":
			if s, ok := AsString(value); ok {
				so.KeyPath = s
			} else {
				log.Printf("некорректный тип для 'crypto_key': %T", value)
			}
		case "trusted_subnet":
			if s, ok := AsString(value); ok {
				so.TrustedSubnet = s
			} else {
				log.Printf("некорректный тип для 'trusted_subnet': %T", value)
			}
		case "grpc_host":
			if s, ok := AsString(value); ok {
				so.GrpcHost = s
			} else {
				log.Printf("некорректный тип для 'grpc_host': %T", value)
			}
		default:
			log.Printf("неизвестный ключ конфига: %s", key)
		}
	}

	return so, nil
}

func getServerOptionsFromFlag() *ServerOptions {
	hostFlag := flag.String("a", "", "Хост")
	storeIntervalFlag := flag.Int("i", 0, "Интервал сохранения данных (сек)")
	restoreFlag := flag.Bool("r", false, "Восстанавливать данные из файла")
	fileStoragePathFlag := flag.String("f", "", "Путь к файлу хранения")
	databaseDSNFlag := flag.String("d", "", "Aдрес подключения к БД")
	keyFlag := flag.String("k", "", "Ключ для хеша")
	auditFileFlag := flag.String("audit-file", "", "Путь к файлу аудита")
	auditURLflag := flag.String("audit-url", "", "URL для аудита")
	bufSizeflag := flag.Int("b", 0, "Размер буфера каналов")
	keyPathFlag := flag.String("crypto-key", "", "Путь до приватного ключа")
	fileConfigFlag := flag.String("c", "", "Путь до файла конфигурации")
	fileConfigFlag = flag.String("config", *fileConfigFlag, "Путь до файла конфигурации")
	trustedSubnetFlag := flag.String("t", "", "Cтроковое представление бесклассовой адресации (CIDR)")
	grpcHostFlag := flag.String("grpc-host", "", "Адрес gRPC сервера")
	flag.Parse()

	so := &ServerOptions{}

	so.Host = *hostFlag
	so.StoreInterval = *storeIntervalFlag
	so.Restore = *restoreFlag
	so.FileStoragePath = *fileStoragePathFlag
	so.DatabaseDSN = *databaseDSNFlag
	so.Key = *keyFlag
	so.AuditFilePath = *auditFileFlag
	so.AuditURL = *auditURLflag
	so.BufferSize = *bufSizeflag
	so.KeyPath = *keyPathFlag
	so.FileConfig = *fileConfigFlag
	so.TrustedSubnet = *trustedSubnetFlag
	so.GrpcHost = *grpcHostFlag

	return so
}

func (so *ServerOptions) CompareAndAddValues(trg *ServerOptions) {
	if so == nil || trg == nil {
		return
	}

	if so.Host == "" {
		so.Host = trg.Host
	}
	if so.FileStoragePath == "" {
		so.FileStoragePath = trg.FileStoragePath
	}
	if so.DatabaseDSN == "" {
		so.DatabaseDSN = trg.DatabaseDSN
	}
	if so.Key == "" {
		so.Key = trg.Key
	}
	if so.AuditFilePath == "" {
		so.AuditFilePath = trg.AuditFilePath
	}
	if so.AuditURL == "" {
		so.AuditURL = trg.AuditURL
	}
	if so.KeyPath == "" {
		so.KeyPath = trg.KeyPath
	}

	if so.TrustedSubnet == "" {
		so.TrustedSubnet = trg.TrustedSubnet
	}

	if so.GrpcHost == "" {
		so.GrpcHost = trg.GrpcHost
	}

	if so.StoreInterval == 0 {
		so.StoreInterval = trg.StoreInterval
	}
	if so.BufferSize == 0 {
		so.BufferSize = trg.BufferSize
	}

}
