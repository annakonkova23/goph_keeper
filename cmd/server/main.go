package main

import (
	"context"
	"log"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"

	"github.com/konkovaanna23/gophkeeper/internal/config"
	"github.com/konkovaanna23/gophkeeper/internal/config/db"
	"go.uber.org/zap"
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

	/*ctx*/
	_, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	quitCh := make(chan os.Signal, 1)
	signal.Notify(quitCh, syscall.SIGQUIT)
	defer signal.Stop(quitCh)

	//errCh := make(chan error, 2)

	cfg := config.GetConfig()
	_ /*database, err */, _ = db.NewConnect(cfg.DSN)
	if err != nil {
		logger.Error("Ошибка при подключении к базе данных:", zap.Error(err))
	} else {
		logger.Info("Подключение к базе данных успешно")
		if err := db.RunMigrations(cfg.DSN); err != nil {
			logger.Error("Ошибка при установке миграций:", zap.Error(err))
			//database = nil
		}
	}

	/*converter := service.NewConverter(ctx, cfg.URLforShort, cfg.FilePath, database, cfg.BufferSize, cfg.BatchSize, cfg.TimeFlushDel)
	server := handler.NewServer(cfg.URLserver, converter, cfg.Key, cfg.AuditFilePath, cfg.AuditURL, cfg.EnableHTTPS, cfg.TrustedSubnet)

	go func() {
		logrus.Printf("Сервер запущен на: %s", cfg.URLserver)
		errCh <- server.Start(ctx)
	}()

	if cfg.GrpcServer != "" {
		go func() {
			grpcServer := grpcserver.NewGrpcServer(converter, cfg.AuditFilePath, cfg.AuditURL)
			errCh <- startGrpcServer(cfg.GrpcServer, grpcServer)
		}()
	}

	for {
		select {
		case <-quitCh:
			fmt.Fprintln(os.Stderr, "Получен сигнал SIGQUIT: goroutine dump (pprof)")

			if p := pprof.Lookup("goroutine"); p != nil {
				err := p.WriteTo(os.Stderr, 2)
				if err != nil {
					logrus.Error(err)
				}
			} else {
				fmt.Fprintln(os.Stderr, "pprof.Lookup(\"goroutine\") вернул nil")
			}
			continue

		case <-ctx.Done():
			logrus.Println("Сервер остановлен")

			shutdownCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			err := pprofSrv.Shutdown(shutdownCtx)
			if err != nil {
				logrus.Error(err)
			}
			return

		case err := <-errCh:
			if err != nil {
				logrus.Fatal("Критическая ошибка в горутине:", err)
			}
		}

	}*/
}

/*
func startGrpcServer(host string, srv *grpcserver.GrpcServer) error {
	listen, err := net.Listen("tcp", host)
	if err != nil {
		return err
	}

	s := grpc.NewServer(grpc.UnaryInterceptor(grpcserver.UnaryInterceptor))

	ss.RegisterShortenerServiceServer(s, srv)

	logrus.Info("сервер gRPC начал работу, host:", host)

	if err := s.Serve(listen); err != nil {
		return err
	}
	return nil
}
*/
