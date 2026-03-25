package main

import (
	"context"
	"fmt"
	"log"
	"net"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"runtime/pprof"
	"syscall"
	"time"

	"github.com/konkovaanna23/gophkeeper/internal/auth"
	"github.com/konkovaanna23/gophkeeper/internal/config"
	"github.com/konkovaanna23/gophkeeper/internal/config/db"
	"github.com/konkovaanna23/gophkeeper/internal/repository"
	"github.com/konkovaanna23/gophkeeper/internal/service"
	"github.com/konkovaanna23/gophkeeper/internal/service/interceptor"
	pb "github.com/konkovaanna23/gophkeeper/pkg/keeperservice"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

var buildVersion string
var buildDate string
var buildCommit string

func main() {

	config.PrintBuildInfo(buildVersion, buildDate, buildCommit)

	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatal(err)
	}
	defer logger.Sync()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	quitCh := make(chan os.Signal, 1)
	signal.Notify(quitCh, syscall.SIGQUIT)
	defer signal.Stop(quitCh)

	errCh := make(chan error, 1)

	cfg := config.GetConfig()
	database, err := db.NewConnect(cfg.DSN)
	if err != nil {
		logger.Fatal("ошибка при подключении к базе данных:", zap.Error(err))
	} else {
		logger.Info("подключение к базе данных успешно")
		if err := db.RunMigrations(cfg.DSN); err != nil {
			logger.Fatal("ошибка при установке миграций:", zap.Error(err))
		}
	}

	repo := repository.NewDBStore(logger, database, time.Duration(cfg.DeleteFileTimeout)*time.Hour)
	jwtManager := auth.NewJWTManager(cfg.KeyAuth, time.Duration(cfg.TokenTTL)*time.Hour)

	key, err := config.GetEncryptionKey(cfg.Key)
	if err != nil {
		logger.Fatal("Некорректно задан ключ", zap.Error(err))
	}
	if cfg.GrpcServer != "" {
		go func() {
			storageServer := service.NewStorageServer(logger, repo, key)
			authServer := service.NewAuthServer(logger, repo, jwtManager)
			errCh <- startGrpcServer(logger, cfg.GrpcServer, authServer, storageServer, auth.UnaryAuthInterceptor(jwtManager), auth.StreamAuthInterceptor(jwtManager))
		}()
	}

	for {
		select {
		case <-quitCh:
			fmt.Fprintln(os.Stderr, "Получен сигнал SIGQUIT: goroutine dump (pprof)")

			if p := pprof.Lookup("goroutine"); p != nil {
				err := p.WriteTo(os.Stderr, 2)
				if err != nil {
					logger.Error("Ошибка записи os.Stderr", zap.Error(err))
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

func startGrpcServer(logger *zap.Logger, host string, authSrv *service.AuthServer, strService *service.StorageServer, interceprtorAuth grpc.UnaryServerInterceptor, interceprtorAuthStream grpc.StreamServerInterceptor) error {
	listen, err := net.Listen("tcp", host)
	if err != nil {
		return err
	}

	unaryChain := grpc.ChainUnaryInterceptor(
		interceprtorAuth,
		interceptor.UnaryLoggerInterceptor(logger),
	)

	streamChain := grpc.ChainStreamInterceptor(
		interceprtorAuthStream,
		interceptor.StreamLoggerInterceptor(logger),
	)

	s := grpc.NewServer(unaryChain,
		streamChain)

	pb.RegisterAuthServiceServer(s, authSrv)
	pb.RegisterStorageServiceServer(s, strService)

	logger.Info("сервер gRPC начал работу", zap.String("host", host))

	if err := s.Serve(listen); err != nil {
		return err
	}
	return nil
}
