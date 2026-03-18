package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"

	"github.com/konkovaanna23/gophkeeper/internal/auth"
	"github.com/konkovaanna23/gophkeeper/internal/crypto"
	"github.com/konkovaanna23/gophkeeper/internal/model"
	"github.com/konkovaanna23/gophkeeper/internal/repository"
	pb "github.com/konkovaanna23/gophkeeper/pkg/keeperservice"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type StorageServer struct {
	pb.UnimplementedAuthServiceServer
	repo repository.StoreRepository
}

func NewStorageService(repo repository.StoreRepository) *StorageServer {
	return &StorageServer{
		repo: repo,
	}
}

func grpcErr(err error) error {
	switch {
	case errors.Is(err, repository.ErrorNotContent):
		return status.Error(codes.NotFound, "not found")
	case errors.Is(err, repository.ErrorVersionConflict):
		return status.Error(codes.Aborted, "version conflict")
	case errors.Is(err, repository.ErrorEmptyUpload):
		return status.Error(codes.FailedPrecondition, "upload has no chunks")
	default:
		return status.Error(codes.Internal, err.Error())
	}
}

func (s *StorageServer) CreateAuthInfo(ctx context.Context, req *pb.AuthInfo) (*pb.CreateInfoResponse, error) {
	userID, ok := ctx.Value(auth.UserIDKey).(string)
	if !ok || userID == "" {
		return nil, status.Error(codes.Unauthenticated, "user is not authenticated")
	}

	metaJSON, _ := json.Marshal(req.GetMeta())
	password, err := crypto.Encrypt([]byte(req.GetPassword()), []byte("key"))
	if err != nil {
		return nil, fmt.Errorf("error crypto password %w", err)
	}
	auth := &model.AuthData{
		ID:       NewUUID(),
		UserID:   userID,
		Login:    req.Login,
		Password: password,
		Meta:     metaJSON,
	}

	version, err := s.repo.CreateAuthData(ctx, auth)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to create item")
	}

	return &pb.CreateInfoResponse{
		Id:      auth.ID,
		Version: version,
	}, nil
}

func (s *StorageServer) GetAuthInfo(ctx context.Context, req *pb.GetInfoRequest) (*pb.StoredAuthInfo, error) {
	userID, ok := ctx.Value(auth.UserIDKey).(string)
	if !ok || userID == "" {
		return nil, status.Error(codes.Unauthenticated, "user is not authenticated")
	}

	auth, err := s.repo.GetAuthData(ctx, userID, req.GetId())
	if err != nil {
		if err == repository.ErrorNotContent {
			return nil, status.Error(codes.NotFound, "item not found")
		}
		return nil, fmt.Errorf("failed to get auth data")
	}

	password, err := crypto.Decrypt(auth.Password, []byte("key"))
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt password")
	}
	var meta map[string]string
	err = json.Unmarshal(auth.Meta, meta)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal meta")
	}
	data := &pb.AuthInfo{
		Login:    auth.Login,
		Password: string(password),
		Meta:     meta,
	}
	return &pb.StoredAuthInfo{
		Id:        auth.ID,
		Data:      data,
		UpdatedAt: timestamppb.New(auth.UpdatedAt),
		Version:   auth.Version,
	}, nil
}

func (s *StorageServer) UpdateAuthInfo(ctx context.Context, req *pb.UpdateAuthInfoRequest) (*pb.UpdateInfoResponse, error) {
	userID, ok := ctx.Value(auth.UserIDKey).(string)
	if !ok || userID == "" {
		return nil, status.Error(codes.Unauthenticated, "user is not authenticated")
	}

	metaJSON, err := json.Marshal(req.GetData().GetMeta())
	if err != nil {
		return nil, fmt.Errorf("failed to marshal meta")
	}
	password, err := crypto.Encrypt([]byte(req.GetData().GetPassword()), []byte("key"))
	if err != nil {
		return nil, fmt.Errorf("error crypto password %w", err)
	}
	auth := &model.AuthData{
		ID:       req.Id,
		UserID:   userID,
		Login:    req.GetData().Login,
		Password: password,
		Meta:     metaJSON,
	}

	newVersion, err := s.repo.UpdateAuthData(ctx, auth, req.GetVersion())
	if err != nil {
		if err == repository.ErrorVersionConflict {
			return nil, status.Error(codes.Aborted, "stale version")
		}
		return nil, status.Error(codes.Internal, "failed to update item")
	}

	return &pb.UpdateInfoResponse{
		NewVersion: newVersion,
	}, nil
}

func (s *StorageServer) DeleteAuthInfo(ctx context.Context, req *pb.DeleteInfoRequest) (*pb.DeleteInfoResponse, error) {
	userID, ok := ctx.Value(auth.UserIDKey).(string)
	if !ok || userID == "" {
		return nil, status.Error(codes.Unauthenticated, "user is not authenticated")
	}

	err := s.repo.DeleteAuthData(ctx, userID, req.GetId(), req.GetVersion())
	if err != nil {
		if err == repository.ErrorVersionConflict {
			return nil, status.Error(codes.Aborted, "stale version")
		}
		return nil, status.Error(codes.Internal, "failed to delete item")
	}

	return &pb.DeleteInfoResponse{}, nil
}

func (s *StorageServer) CreateTextInfo(ctx context.Context, req *pb.TextInfo) (*pb.CreateInfoResponse, error) {
	userID, ok := ctx.Value(auth.UserIDKey).(string)
	if !ok || userID == "" {
		return nil, status.Error(codes.Unauthenticated, "user is not authenticated")
	}

	metaJSON, _ := json.Marshal(req.GetMeta())

	text := &model.TextData{
		ID:     NewUUID(),
		UserID: userID,
		Data:   req.GetText(),
		Meta:   metaJSON,
	}

	version, err := s.repo.CreateTextData(ctx, text)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to create item")
	}

	return &pb.CreateInfoResponse{
		Id:      text.ID,
		Version: version,
	}, nil
}

func (s *StorageServer) GetTextInfo(ctx context.Context, req *pb.GetInfoRequest) (*pb.StoredTextInfo, error) {
	userID, ok := ctx.Value(auth.UserIDKey).(string)
	if !ok || userID == "" {
		return nil, status.Error(codes.Unauthenticated, "user is not authenticated")
	}

	text, err := s.repo.GetTextData(ctx, userID, req.GetId())
	if err != nil {
		return nil, grpcErr(err)
	}

	var meta map[string]string
	err = json.Unmarshal(text.Meta, meta)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal meta")
	}
	data := &pb.TextInfo{
		Text: text.Data,
		Meta: meta,
	}
	return &pb.StoredTextInfo{
		Id:        text.ID,
		Data:      data,
		UpdatedAt: timestamppb.New(text.UpdatedAt),
		Version:   text.Version,
	}, nil
}

func (s *StorageServer) UpdateTextInfo(ctx context.Context, req *pb.UpdateTextInfoRequest) (*pb.UpdateInfoResponse, error) {
	userID, ok := ctx.Value(auth.UserIDKey).(string)
	if !ok || userID == "" {
		return nil, status.Error(codes.Unauthenticated, "user is not authenticated")
	}

	metaJSON, err := json.Marshal(req.GetData().GetMeta())
	if err != nil {
		return nil, fmt.Errorf("failed to marshal meta")
	}

	text := &model.TextData{
		ID:     req.Id,
		UserID: userID,
		Data:   req.GetData().GetText(),
		Meta:   metaJSON,
	}

	newVersion, err := s.repo.UpdateTextData(ctx, text, req.GetVersion())
	if err != nil {
		return nil, grpcErr(err)
	}

	return &pb.UpdateInfoResponse{
		NewVersion: newVersion,
	}, nil
}

func (s *StorageServer) DeleteTextInfo(ctx context.Context, req *pb.DeleteInfoRequest) (*pb.DeleteInfoResponse, error) {
	userID, ok := ctx.Value(auth.UserIDKey).(string)
	if !ok || userID == "" {
		return nil, status.Error(codes.Unauthenticated, "user is not authenticated")
	}

	err := s.repo.DeleteTextData(ctx, userID, req.GetId(), req.GetVersion())
	if err != nil {
		return nil, grpcErr(err)
	}

	return &pb.DeleteInfoResponse{}, nil
}

func (s *StorageServer) StartUpload(ctx context.Context, req *pb.StartUploadRequest) (*pb.StartUploadResponse, error) {
	userID, ok := ctx.Value(auth.UserIDKey).(string)
	if !ok || userID == "" {
		return nil, status.Error(codes.Unauthenticated, "user is not authenticated")
	}

	if req.GetFilename() == "" {
		return nil, status.Error(codes.InvalidArgument, "filename is required")
	}

	uploadID, fileID, err := s.repo.StartUpload(
		ctx,
		userID,
		req.GetFileId(),
		req.GetExpectedVersion(),
		req.GetFilename(),
		req.GetMeta(),
		NewUUID,
	)
	if err != nil {
		return nil, grpcErr(err)
	}

	return &pb.StartUploadResponse{
		UploadId: uploadID,
		FileId:   fileID,
	}, nil
}

func (s *StorageServer) UploadChunks(stream pb.StorageService_UploadChunksServer) error {
	userID, ok := stream.Context().Value(auth.UserIDKey).(string)
	if !ok || userID == "" {
		return status.Error(codes.Unauthenticated, "user is not authenticated")
	}

	var uploadID string
	var received int64
	var totalBytes int64

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&pb.UploadChunksResponse{
				UploadId:       uploadID,
				ReceivedChunks: received,
				TotalBytes:     totalBytes,
			})
		}
		if err != nil {
			return status.Error(codes.Internal, "failed to receive stream message")
		}

		if req.GetUploadId() == "" {
			return status.Error(codes.InvalidArgument, "upload_id is required")
		}
		if uploadID == "" {
			uploadID = req.GetUploadId()
		}
		if req.GetUploadId() != uploadID {
			return status.Error(codes.InvalidArgument, "all chunks must belong to one upload_id")
		}
		if len(req.GetData()) == 0 {
			return status.Error(codes.InvalidArgument, "chunk data is empty")
		}

		if err := s.repo.PutUploadChunk(
			stream.Context(),
			userID,
			req.GetUploadId(),
			req.GetChunkNo(),
			req.GetData(),
		); err != nil {
			return grpcErr(err)
		}

		received++
		totalBytes += int64(len(req.GetData()))
	}
}

func (s *StorageServer) CommitUpload(ctx context.Context, req *pb.CommitUploadRequest) (*pb.CommitUploadResponse, error) {
	userID, ok := ctx.Value(auth.UserIDKey).(string)
	if !ok || userID == "" {
		return nil, status.Error(codes.Unauthenticated, "user is not authenticated")
	}

	if req.GetUploadId() == "" {
		return nil, status.Error(codes.InvalidArgument, "upload_id is required")
	}

	fileID, newVersion, err := s.repo.CommitUpload(
		ctx,
		userID,
		req.GetUploadId(),
		req.GetFileId(),
		req.GetExpectedVersion(),
		req.GetChecksum(),
	)
	if err != nil {
		return nil, grpcErr(err)
	}

	return &pb.CommitUploadResponse{
		FileId:     fileID,
		NewVersion: newVersion,
	}, nil
}

func (s *StorageServer) GetFileMeta(ctx context.Context, req *pb.GetFileMetaRequest) (*pb.GetFileMetaResponse, error) {
	userID, ok := ctx.Value(auth.UserIDKey).(string)
	if !ok || userID == "" {
		return nil, status.Error(codes.Unauthenticated, "user is not authenticated")
	}

	meta, err := s.repo.GetFileMeta(ctx, userID, req.GetFileId())
	if err != nil {
		return nil, grpcErr(err)
	}

	outMeta := map[string]string{}
	_ = json.Unmarshal(meta.MetaJSON, &outMeta)

	return &pb.GetFileMetaResponse{
		File: &pb.FileMeta{
			FileId:         meta.FileID,
			CurrentVersion: meta.CurrentVersion,
			Filename:       meta.Filename,
			SizeBytes:      meta.SizeBytes,
			Checksum:       meta.Checksum,
			Meta:           outMeta,
		},
	}, nil
}

func (s *StorageServer) DownloadFile(req *pb.DownloadFileRequest, stream pb.StorageService_DownloadFileServer) error {
	userID, ok := stream.Context().Value(auth.UserIDKey).(string)
	if !ok || userID == "" {
		return status.Error(codes.Unauthenticated, "user is not authenticated")
	}

	err := s.repo.StreamFileChunks(
		stream.Context(),
		userID,
		req.GetFileId(),
		req.GetVersion(),
		func(chunkNo int64, data []byte) error {
			return stream.Send(&pb.DownloadFileChunk{
				ChunkNo: chunkNo,
				Data:    data,
			})
		},
	)
	if err != nil {
		return grpcErr(err)
	}

	return nil
}

func (s *StorageServer) DeleteFile(ctx context.Context, req *pb.DeleteFileRequest) (*pb.DeleteInfoResponse, error) {
	userID, ok := ctx.Value(auth.UserIDKey).(string)
	if !ok || userID == "" {
		return nil, status.Error(codes.Unauthenticated, "user is not authenticated")
	}

	err := s.repo.DeleteFile(ctx, userID, req.GetFileId(), req.GetExpectedVersion())
	if err != nil {
		return nil, grpcErr(err)
	}

	return &pb.DeleteInfoResponse{}, nil
}

func (s *StorageServer) CreateBankCardDetailsInfo(ctx context.Context, req *pb.BankCardDetails) (*pb.CreateInfoResponse, error) {
	userID, ok := ctx.Value(auth.UserIDKey).(string)
	if !ok || userID == "" {
		return nil, status.Error(codes.Unauthenticated, "user is not authenticated")
	}

	metaJSON, _ := json.Marshal(req.GetMeta())
	number, err := crypto.Encrypt([]byte(req.GetNumber()), []byte("key"))
	if err != nil {
		return nil, fmt.Errorf("error crypto number %w", err)
	}

	last4 := ""
	if len(req.GetNumber()) >= 4 {
		last4 = req.GetNumber()[len(req.GetNumber())-4:]
	}
	last4n, _ := strconv.ParseUint(last4, 10, 32)

	card := &model.BankCardData{
		ID:              NewUUID(),
		UserID:          userID,
		Last4:           uint32(last4n),
		NumberEncrypted: number,
		ExpMonth:        req.GetExpMonth(),
		ExpYear:         req.GetExpYear(),
		Owner:           req.GetOwner(),
		Meta:            metaJSON,
	}

	version, err := s.repo.CreateBankCardData(ctx, card)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to create item")
	}

	return &pb.CreateInfoResponse{
		Id:      card.ID,
		Version: version,
	}, nil
}

func (s *StorageServer) GetBankCardDetails(ctx context.Context, req *pb.GetInfoRequest) (*pb.StoredBankCardDetails, error) {
	userID, ok := ctx.Value(auth.UserIDKey).(string)
	if !ok || userID == "" {
		return nil, status.Error(codes.Unauthenticated, "user is not authenticated")
	}

	card, err := s.repo.GetBankCardData(ctx, userID, req.GetId())
	if err != nil {
		return nil, grpcErr(err)
	}

	number, err := crypto.Decrypt(card.NumberEncrypted, []byte("key"))
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt password")
	}
	var meta map[string]string
	err = json.Unmarshal(card.Meta, meta)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal meta")
	}
	data := &pb.BankCardDetails{
		Number:   string(number),
		ExpMonth: card.ExpMonth,
		ExpYear:  card.ExpYear,
		Owner:    card.Owner,
		Meta:     meta,
	}
	return &pb.StoredBankCardDetails{
		Id:        card.ID,
		Data:      data,
		UpdatedAt: timestamppb.New(card.UpdatedAt),
		Version:   card.Version,
	}, nil
}

func (s *StorageServer) UpdateBankCardDetails(ctx context.Context, req *pb.UpdateBankCardDetailsRequest) (*pb.UpdateInfoResponse, error) {
	userID, ok := ctx.Value(auth.UserIDKey).(string)
	if !ok || userID == "" {
		return nil, status.Error(codes.Unauthenticated, "user is not authenticated")
	}

	metaJSON, err := json.Marshal(req.GetData().GetMeta())
	if err != nil {
		return nil, fmt.Errorf("failed to marshal meta")
	}
	number, err := crypto.Encrypt([]byte(req.GetData().GetNumber()), []byte("key"))
	if err != nil {
		return nil, fmt.Errorf("error crypto password %w", err)
	}

	last4 := ""
	if len(req.GetData().GetNumber()) >= 4 {
		last4 = req.GetData().GetNumber()[len(req.GetData().GetNumber())-4:]
	}
	last4n, _ := strconv.ParseUint(last4, 10, 32)

	card := &model.BankCardData{
		ID:              req.Id,
		UserID:          userID,
		Last4:           uint32(last4n),
		NumberEncrypted: number,
		ExpMonth:        req.GetData().GetExpMonth(),
		ExpYear:         req.GetData().GetExpYear(),
		Owner:           req.GetData().GetOwner(),
		Meta:            metaJSON,
	}

	newVersion, err := s.repo.UpdateBankCardData(ctx, card, req.GetVersion())
	if err != nil {
		return nil, grpcErr(err)
	}

	return &pb.UpdateInfoResponse{
		NewVersion: newVersion,
	}, nil
}

func (s *StorageServer) DeleteBankCardDetails(ctx context.Context, req *pb.DeleteInfoRequest) (*pb.DeleteInfoResponse, error) {
	userID, ok := ctx.Value(auth.UserIDKey).(string)
	if !ok || userID == "" {
		return nil, status.Error(codes.Unauthenticated, "user is not authenticated")
	}

	err := s.repo.DeleteBankCardData(ctx, userID, req.GetId(), req.GetVersion())
	if err != nil {
		if err == repository.ErrorVersionConflict {
			return nil, status.Error(codes.Aborted, "stale version")
		}
		return nil, status.Error(codes.Internal, "failed to delete item")
	}

	return &pb.DeleteInfoResponse{}, nil
}
