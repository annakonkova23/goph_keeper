package main

import (
	"context"
	"fmt"
	"log"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"runtime/pprof"
	"syscall"

	"github.com/konkovaanna23/gophkeeper/internal/config"
	"github.com/konkovaanna23/gophkeeper/internal/handler"
	"go.uber.org/zap"
)

var buildVersion string
var buildDate string
var buildCommit string

// @title           GophKeeper API
// @version         1.0
// @description     API для безопасного хранения паролей, текстов, карт и файлов.
// @termsOfService  http://swagger.io/terms/

// @license.name  MIT
// @license.url   https://opensource.org/licenses/MIT

// @host      localhost:8080
// @BasePath  /api

// @securityDefinitions.apikey  ApiKeyAuth
// @in                          header
// @name                        Authorization
// @description                 Bearer Token: "Bearer {token}"
func main() {

	config.PrintBuildInfo(buildVersion, buildDate, buildCommit)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	quitCh := make(chan os.Signal, 1)
	signal.Notify(quitCh, syscall.SIGQUIT)
	defer signal.Stop(quitCh)

	errCh := make(chan error, 1)

	cfg := config.GetConfigClient()

	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatal(err)
	}
	defer logger.Sync()

	server, err := handler.NewServer(logger, cfg.Host, cfg.GrpcServer)
	if err != nil {
		logger.Fatal("ошибка создания экземпляра сервера", zap.Error(err))
	}

	go func() {
		logger.Info("Сервер запущен", zap.String("host", cfg.Host))
		errCh <- server.Start(ctx)
	}()

	for {
		select {
		case <-quitCh:
			fmt.Fprintln(os.Stderr, "Получен сигнал SIGQUIT: goroutine dump (pprof)")

			if p := pprof.Lookup("goroutine"); p != nil {
				err := p.WriteTo(os.Stderr, 2)
				if err != nil {
					logger.Error("Ошибка записи в Stderr", zap.Error(err))
				}
			} else {
				fmt.Fprintln(os.Stderr, "pprof.Lookup(\"goroutine\") вернул nil")
			}
			continue

		case <-ctx.Done():
			logger.Info("Сервер остановлен")
			return
		case err := <-errCh:
			if err != nil {
				logger.Fatal("Критическая ошибка в горутине:", zap.Error(err))
			}
		}

	}
}
