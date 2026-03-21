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

// Берём безопасный размер куска.
// Обычно дефолтный max receive у gRPC ~4MB, поэтому 1MB — нормальный запас.
const defaultUploadChunkSize = 1 * 1024 * 1024 // 1 MiB

type StorageServiceClient struct {
	conn   *grpc.ClientConn
	client pb.StorageServiceClient
}

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

func (c *StorageServiceClient) Close() error {
	if c == nil || c.conn == nil {
		return nil
	}
	return c.conn.Close()
}

// ---------- AuthInfo ----------

func (c *StorageServiceClient) CreateAuthInfo(ctx context.Context, in *pb.AuthInfo) (*pb.CreateInfoResponse, error) {
	return c.client.CreateAuthInfo(ctx, in)
}

func (c *StorageServiceClient) GetAuthInfo(ctx context.Context, in *pb.GetInfoRequest) (*pb.StoredAuthInfo, error) {
	return c.client.GetAuthInfo(ctx, in)
}

func (c *StorageServiceClient) UpdateAuthInfo(ctx context.Context, in *pb.UpdateAuthInfoRequest) (*pb.UpdateInfoResponse, error) {
	return c.client.UpdateAuthInfo(ctx, in)
}

func (c *StorageServiceClient) DeleteAuthInfo(ctx context.Context, in *pb.DeleteInfoRequest) (*pb.DeleteInfoResponse, error) {
	return c.client.DeleteAuthInfo(ctx, in)
}

// ---------- TextInfo ----------

func (c *StorageServiceClient) CreateTextInfo(ctx context.Context, in *pb.TextInfo) (*pb.CreateInfoResponse, error) {
	return c.client.CreateTextInfo(ctx, in)
}

func (c *StorageServiceClient) GetTextInfo(ctx context.Context, in *pb.GetInfoRequest) (*pb.StoredTextInfo, error) {
	return c.client.GetTextInfo(ctx, in)
}

func (c *StorageServiceClient) UpdateTextInfo(ctx context.Context, in *pb.UpdateTextInfoRequest) (*pb.UpdateInfoResponse, error) {
	return c.client.UpdateTextInfo(ctx, in)
}

func (c *StorageServiceClient) DeleteTextInfo(ctx context.Context, in *pb.DeleteInfoRequest) (*pb.DeleteInfoResponse, error) {
	return c.client.DeleteTextInfo(ctx, in)
}

// ---------- BankCardDetails ----------

func (c *StorageServiceClient) CreateBankCardDetails(ctx context.Context, in *pb.BankCardDetails) (*pb.CreateInfoResponse, error) {
	return c.client.CreateBankCardDetails(ctx, in)
}

func (c *StorageServiceClient) GetBankCardDetails(ctx context.Context, in *pb.GetInfoRequest) (*pb.StoredBankCardDetails, error) {
	return c.client.GetBankCardDetails(ctx, in)
}

func (c *StorageServiceClient) UpdateBankCardDetails(ctx context.Context, in *pb.UpdateBankCardDetailsRequest) (*pb.UpdateInfoResponse, error) {
	return c.client.UpdateBankCardDetails(ctx, in)
}

func (c *StorageServiceClient) DeleteBankCardDetails(ctx context.Context, in *pb.DeleteInfoRequest) (*pb.DeleteInfoResponse, error) {
	return c.client.DeleteBankCardDetails(ctx, in)
}

// ---------- File API ----------

// CreateFileFromPath:
// - принимает путь к локальному файлу
// - вызывает StartUpload
// - режет файл на куски
// - отправляет UploadChunks
// - делает CommitUpload
// - возвращает file_id и new_version
func (c *StorageServiceClient) CreateFileFromPath(ctx context.Context, filePath string, meta map[string]string) (*pb.CommitUploadResponse, error) {
	return c.uploadFileFromPath(ctx, filePath, meta, 0, false)
}

// UpdateFileFromPath:
// - принимает путь к локальному файлу и текущую версию
// - делает тот же цикл, но с version
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

// DownloadFileToPath:
// - принимает путь к папке, имя файла и версию
// - вызывает GetFileMeta
// - вызывает DownloadFile
// - формирует файл по пути path/filename
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
