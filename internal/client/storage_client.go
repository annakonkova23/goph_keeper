// Package client предоставляет клиент для gRPC-сервиса GophKeeper.
//
// Клиент позволяет:
//   - Управлять данными: логины, тексты, банковские карты.
//   - Загружать, скачивать и обновлять файлы (с потоковой передачей).
//   - Работать с версионированием (оптимистичные блокировки).
//
// Особенности:
//   - Автоматическое дробление файлов на чанки (1 MiB).
//   - Расчёт SHA256 при загрузке.
//
// Пример использования:
//
//	client, err := client.NewStorageServiceClient("localhost:50051")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer client.Close()
//
//	// Загрузка файла
//	resp, err := client.CreateFileFromPath(ctx, "/home/user/photo.jpg", map[string]string{"note": "avatar"}, 0)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Файл загружен: ID=%s, версия=%d\n", resp.FileId, resp.NewVersion)
//
//	// Скачивание
//	meta, path, err := client.DownloadFileToPath(ctx, "/tmp/download.jpg", resp.FileId, resp.NewVersion)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Файл сохранён: %s, размер=%d\n", path, meta.File.SizeBytes)
package client

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"

	pb "github.com/konkovaanna23/gophkeeper/pkg/keeperservice"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Размер куска на который разделяем файл.
const defaultUploadChunkSize = 1 * 1024 * 1024 // 1 MiB

// StorageServiceClient — gRPC-клиент для взаимодействия с StorageService.
//
// Предоставляет методы для:
//   - CRUD операций с данными (логины, тексты, карты).
//   - Загрузки/скачивания файлов с автоматическим чанкингом.
type StorageServiceClient struct {
	conn   *grpc.ClientConn
	client pb.StorageServiceClient
}

// NewStorageServiceClient создаёт новый клиент для StorageService.
//
// Подключается к gRPC-серверу по указанному адресу.
// Использует незащищённое соединение (insecure) — подходит для разработки.
// В продакшене рекомендуется использовать TLS.
//
// Возвращает:
//   - *StorageServiceClient: готовый клиент.
//   - ошибка, если не удалось установить соединение.
//
// Пример:
//
//	client, err := client.NewStorageServiceClient("localhost:50051")
//	if err != nil {
//	    log.Fatal("Не удалось подключиться:", err)
//	}
//	defer client.Close()
func NewStorageServiceClient(grpcAddr string) (*StorageServiceClient, error) {
	conn, err := grpc.NewClient(
		grpcAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}

	return &StorageServiceClient{
		conn:   conn,
		client: pb.NewStorageServiceClient(conn),
	}, nil
}

// Close закрывает gRPC-соединение.
// Должен использоваться с defer:
//
//	defer client.Close()
func (c *StorageServiceClient) Close() error {
	if c == nil || c.conn == nil {
		return nil
	}
	return c.conn.Close()
}

// ---------- AuthInfo ----------
// CreateAuthInfo сохраняет учётные данные (логин/пароль).
func (c *StorageServiceClient) CreateAuthInfo(ctx context.Context, in *pb.AuthInfo) (*pb.CreateInfoResponse, error) {
	return c.client.CreateAuthInfo(ctx, in)
}

// GetAuthInfo получает сохранённые данные по ID.
func (c *StorageServiceClient) GetAuthInfo(ctx context.Context, in *pb.GetInfoRequest) (*pb.StoredAuthInfo, error) {
	return c.client.GetAuthInfo(ctx, in)
}

// UpdateAuthInfo обновляет запись с проверкой версии.
func (c *StorageServiceClient) UpdateAuthInfo(ctx context.Context, in *pb.UpdateAuthInfoRequest) (*pb.UpdateInfoResponse, error) {
	return c.client.UpdateAuthInfo(ctx, in)
}

// DeleteAuthInfo удаляет запись по ID и версии.
func (c *StorageServiceClient) DeleteAuthInfo(ctx context.Context, in *pb.DeleteInfoRequest) (*pb.DeleteInfoResponse, error) {
	return c.client.DeleteAuthInfo(ctx, in)
}

// ---------- TextInfo ----------
// CreateTextInfo сохраняет текстовую заметку.
func (c *StorageServiceClient) CreateTextInfo(ctx context.Context, in *pb.TextInfo) (*pb.CreateInfoResponse, error) {
	return c.client.CreateTextInfo(ctx, in)
}

// GetTextInfo получает текст по ID.
func (c *StorageServiceClient) GetTextInfo(ctx context.Context, in *pb.GetInfoRequest) (*pb.StoredTextInfo, error) {
	return c.client.GetTextInfo(ctx, in)
}

// UpdateTextInfo обновляет текстовую запись.
func (c *StorageServiceClient) UpdateTextInfo(ctx context.Context, in *pb.UpdateTextInfoRequest) (*pb.UpdateInfoResponse, error) {
	return c.client.UpdateTextInfo(ctx, in)
}

// DeleteTextInfo удаляет текстовую запись.
func (c *StorageServiceClient) DeleteTextInfo(ctx context.Context, in *pb.DeleteInfoRequest) (*pb.DeleteInfoResponse, error) {
	return c.client.DeleteTextInfo(ctx, in)
}

// ---------- BankCardDetails ----------
// CreateBankCardDetails сохраняет данные карты.
func (c *StorageServiceClient) CreateBankCardDetails(ctx context.Context, in *pb.BankCardDetails) (*pb.CreateInfoResponse, error) {
	return c.client.CreateBankCardDetails(ctx, in)
}

// GetBankCardDetails получает данные карты по ID.
func (c *StorageServiceClient) GetBankCardDetails(ctx context.Context, in *pb.GetInfoRequest) (*pb.StoredBankCardDetails, error) {
	return c.client.GetBankCardDetails(ctx, in)
}

// UpdateBankCardDetails обновляет данные карты.
func (c *StorageServiceClient) UpdateBankCardDetails(ctx context.Context, in *pb.UpdateBankCardDetailsRequest) (*pb.UpdateInfoResponse, error) {
	return c.client.UpdateBankCardDetails(ctx, in)
}

// DeleteBankCardDetails удаляет запись о карте.
func (c *StorageServiceClient) DeleteBankCardDetails(ctx context.Context, in *pb.DeleteInfoRequest) (*pb.DeleteInfoResponse, error) {
	return c.client.DeleteBankCardDetails(ctx, in)
}

// ---------- File API ----------

// CreateFileFromPath загружает файл на сервер.
//
// Этапы:
//  1. Открывает локальный файл.
//  2. Вызывает StartUpload (создаёт сессию).
//  3. Делит файл на чанки по 1 MiB.
//  4. Отправляет чанки через UploadChunks(stream).
//  5. Считает SHA256.
//  6. Финализирует через CommitUpload.
//
// Параметры:
//   - ctx: контекст.
//   - filePath: путь к локальному файлу.
//   - meta: метаданные (например, "тип", "примечание").
//
// Возвращает:
//   - CommitUploadResponse: file_id и new_version.
//   - ошибка при чтении, отправке или верификации.
//
// Пример:
//
//	resp, err := client.CreateFileFromPath(ctx, "/home/user/doc.pdf", map[string]string{"cat": "work"})
func (c *StorageServiceClient) CreateFileFromPath(ctx context.Context, filePath string, meta map[string]string) (*pb.CommitUploadResponse, error) {
	return c.uploadFileFromPath(ctx, filePath, meta, 0, false)
}

// UpdateFileFromPath обновляет существующий файл.
//
// Аналогичен CreateFileFromPath, но:
//   - Ожидает текущую версию.
//   - Передаёт version в StartUpload и CommitUpload.
//
// Защита от перезаписи старой версии.
//
// Возвращает новую версию файла.
func (c *StorageServiceClient) UpdateFileFromPath(ctx context.Context, filePath string, meta map[string]string, version int64) (*pb.CommitUploadResponse, error) {
	return c.uploadFileFromPath(ctx, filePath, meta, version, true)
}

func (c *StorageServiceClient) uploadFileFromPath(ctx context.Context, filePath string, meta map[string]string, version int64, isUpdate bool) (*pb.CommitUploadResponse, error) {

	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("ошибка открытия файла: %w", err)
	}
	defer f.Close()

	filename := filepath.Base(filePath)

	startReq := &pb.StartUploadRequest{
		Filename: filename,
		Meta:     meta,
	}

	if isUpdate {
		startReq.Version = version
	}

	startResp, err := c.client.StartUpload(ctx, startReq)
	if err != nil {
		return nil, err
	}

	stream, err := c.client.UploadChunks(ctx)
	if err != nil {
		return nil, err
	}

	buf := make([]byte, defaultUploadChunkSize)
	hasher := sha256.New()
	var chunkNo int64

	for {
		n, readErr := f.Read(buf)
		if n > 0 {
			chunkNo++
			part := append([]byte(nil), buf[:n]...)

			if _, err := hasher.Write(part); err != nil {
				return nil, fmt.Errorf("hash chunk: %w", err)
			}

			if err := stream.Send(&pb.UploadChunkRequest{
				UploadId: startResp.UploadId,
				ChunkNo:  chunkNo,
				Data:     part,
			}); err != nil {
				return nil, fmt.Errorf("send chunk #%d: %w", chunkNo, err)
			}
		}

		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return nil, readErr
		}
	}

	uploadResp, err := stream.CloseAndRecv()
	if err != nil {
		return nil, err
	}
	_ = uploadResp

	checksum := hex.EncodeToString(hasher.Sum(nil))

	commitReq := &pb.CommitUploadRequest{
		UploadId: startResp.UploadId,
		FileId:   startResp.FileId,
		Checksum: checksum,
	}
	if isUpdate {
		commitReq.Version = version
	}

	commitResp, err := c.client.CommitUpload(ctx, commitReq)
	if err != nil {
		return nil, err
	}

	return commitResp, nil
}

// DownloadFileToPath скачивает файл и сохраняет его локально.
//
// Этапы:
//  1. Получает метаданные через GetFileMeta.
//  2. Открывает локальный файл для записи.
//  3. Получает чанки через DownloadFile(stream).
//  4. Записывает в файл.
//
// Параметры:
//   - ctx: контекст.
//   - dirPath: полный путь к файлу (включая имя).
//   - fileid: ID файла на сервере.
//   - version: версия файла (0 = последняя).
//
// Возвращает:
//   - GetFileMetaResponse: метаданные файла.
//   - путь к сохранённому файлу.
//   - ошибка при получении или записи.
//
// Пример:
//
//	meta, path, err := client.DownloadFileToPath(ctx, "/tmp/photo.jpg", "file-123", 0)
func (c *StorageServiceClient) DownloadFileToPath(ctx context.Context, dirPath string, fileid string, version int64) (*pb.GetFileMetaResponse, string, error) {

	metaResp, err := c.client.GetFileMeta(ctx, &pb.GetFileMetaRequest{
		FileId: fileid,
	})
	if err != nil {
		return nil, "", err
	}

	out, err := os.Create(dirPath)
	if err != nil {
		return nil, "", fmt.Errorf("ошибка создания файла: %w", err)
	}
	defer out.Close()

	stream, err := c.client.DownloadFile(ctx, &pb.DownloadFileRequest{
		FileId:  fileid,
		Version: version,
	})
	if err != nil {
		return nil, "", err
	}

	for {
		chunk, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, "", err
		}

		if _, err := out.Write(chunk.Data); err != nil {
			return nil, "", fmt.Errorf("ошибка записи куска в файл: %w", err)
		}
	}

	return metaResp, dirPath, nil
}

// DeleteFile удаляет файл по ID и версии.
//
// Требует точной версии для защиты от гонок.
//
// Возвращает ошибку, если:
//   - файл не существует,
//   - версия не совпадает,
//   - нет доступа.
func (c *StorageServiceClient) DeleteFile(ctx context.Context, fileID string, version int64) error {

	_, err := c.client.DeleteFile(ctx, &pb.DeleteFileRequest{
		FileId:  fileID,
		Version: version,
	})
	if err != nil {
		return err
	}

	return nil
}
