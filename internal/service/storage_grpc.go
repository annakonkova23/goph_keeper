package service

import (
	"context"
	"fmt"
	"github.com/konkovaanna23/gophkeeper/internal/crypto"
	"github.com/konkovaanna23/gophkeeper/internal/model"
	pb "github.com/konkovaanna23/gophkeeper/pkg/keeperservice"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"io"
)

func (s *StorageServer) CreateAuthInfo(ctx context.Context, req *pb.AuthInfo) (*pb.CreateInfoResponse, error) {
	userID, err := userIDFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	metaJSON, err := marshalMeta(req.Meta)
	if err != nil {
		return nil, err
	}

	password, err := crypto.Encrypt([]byte(req.GetPassword()), s.key)
	if err != nil {
		s.lgr.Error("ошибка шифрования пароля", zap.Error(err))
		return nil, fmt.Errorf("ошибка шифрования пароля %w", err)
	}

	auth := &model.AuthData{
		ID:       NewUUID(),
		UserID:   userID,
		Login:    req.Login,
		Password: password,
		Meta:     metaJSON,
	}

	version, err := s.repo.Auths().Create(ctx, auth)
	if err != nil {
		s.lgr.Error("ошибка создания auth информации", zap.Error(err))
		return nil, grpcErr(err)
	}

	return &pb.CreateInfoResponse{
		Id:      auth.ID,
		Version: version,
	}, nil
}

func (s *StorageServer) GetAuthInfo(ctx context.Context, req *pb.GetInfoRequest) (*pb.StoredAuthInfo, error) {
	userID, err := userIDFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	auth, err := s.repo.Auths().Get(ctx, userID, req.GetId())
	if err != nil {
		s.lgr.Error("ошибка получения auth информации", zap.Error(err))
		return nil, grpcErr(err)
	}

	password, err := crypto.Decrypt(auth.Password, s.key)
	if err != nil {
		s.lgr.Error("ошибка дешифрования пароля", zap.Error(err))
		return nil, fmt.Errorf("ошибка дешифрования пароля")
	}

	meta, err := unmarshalMeta(auth.Meta)
	if err != nil {
		return nil, err
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
	userID, err := userIDFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	metaJSON, err := marshalMeta(req.GetData().GetMeta())
	if err != nil {
		return nil, err
	}

	password, err := crypto.Encrypt([]byte(req.GetData().GetPassword()), s.key)
	if err != nil {
		s.lgr.Error("ошибка шифрования пароля", zap.Error(err))
		return nil, fmt.Errorf("ошибка шифрования пароля %w", err)
	}

	auth := &model.AuthData{
		ID:       req.Id,
		UserID:   userID,
		Login:    req.GetData().Login,
		Password: password,
		Meta:     metaJSON,
	}

	newVersion, err := s.repo.Auths().Update(ctx, auth, req.GetVersion())
	if err != nil {
		s.lgr.Error("ошибка обновления auth", zap.Error(err))
		return nil, grpcErr(err)
	}

	return &pb.UpdateInfoResponse{
		NewVersion: newVersion,
	}, nil
}

func (s *StorageServer) DeleteAuthInfo(ctx context.Context, req *pb.DeleteInfoRequest) (*pb.DeleteInfoResponse, error) {
	userID, err := userIDFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	err = s.repo.Auths().Delete(ctx, userID, req.GetId(), req.GetVersion())
	if err != nil {
		s.lgr.Error("ошибка удаления auth", zap.Error(err))
		return nil, grpcErr(err)
	}

	return &pb.DeleteInfoResponse{}, nil
}

func (s *StorageServer) CreateTextInfo(ctx context.Context, req *pb.TextInfo) (*pb.CreateInfoResponse, error) {
	userID, err := userIDFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	metaJSON, err := marshalMeta(req.GetMeta())
	if err != nil {
		return nil, err
	}

	text := &model.TextData{
		ID:     NewUUID(),
		UserID: userID,
		Data:   req.GetText(),
		Meta:   metaJSON,
	}

	version, err := s.repo.Texts().Create(ctx, text)
	if err != nil {
		s.lgr.Error("ошибка создания text информации", zap.Error(err))
		return nil, grpcErr(err)
	}

	return &pb.CreateInfoResponse{
		Id:      text.ID,
		Version: version,
	}, nil
}

func (s *StorageServer) GetTextInfo(ctx context.Context, req *pb.GetInfoRequest) (*pb.StoredTextInfo, error) {
	userID, err := userIDFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	text, err := s.repo.Texts().Get(ctx, userID, req.GetId())
	if err != nil {
		return nil, grpcErr(err)
	}

	meta, err := unmarshalMeta(text.Meta)
	if err != nil {
		return nil, err
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
	userID, err := userIDFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	metaJSON, err := marshalMeta(req.GetData().GetMeta())
	if err != nil {
		return nil, err
	}

	text := &model.TextData{
		ID:     req.Id,
		UserID: userID,
		Data:   req.GetData().GetText(),
		Meta:   metaJSON,
	}

	newVersion, err := s.repo.Texts().Update(ctx, text, req.GetVersion())
	if err != nil {
		s.lgr.Error("ошибка обновления text информации", zap.Error(err))
		return nil, grpcErr(err)
	}

	return &pb.UpdateInfoResponse{
		NewVersion: newVersion,
	}, nil
}

func (s *StorageServer) DeleteTextInfo(ctx context.Context, req *pb.DeleteInfoRequest) (*pb.DeleteInfoResponse, error) {
	userID, err := userIDFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	err = s.repo.Texts().Delete(ctx, userID, req.GetId(), req.GetVersion())
	if err != nil {
		s.lgr.Error("ошибка удаления text информации", zap.Error(err))
		return nil, grpcErr(err)
	}

	return &pb.DeleteInfoResponse{}, nil
}

func (s *StorageServer) CreateBankCardDetails(ctx context.Context, req *pb.BankCardDetails) (*pb.CreateInfoResponse, error) {
	userID, err := userIDFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	metaJSON, err := marshalMeta(req.GetMeta())
	if err != nil {
		return nil, err
	}

	number, err := crypto.Encrypt([]byte(req.GetNumber()), s.key)
	if err != nil {
		s.lgr.Error("ошибка шифрования номера карты", zap.Error(err))
		return nil, fmt.Errorf("ошибка шифрования номера карты %w", err)
	}

	last4n, err := parseLast4(req.Number)
	if err != nil {
		return nil, err
	}

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

	version, err := s.repo.Cards().Create(ctx, card)
	if err != nil {
		s.lgr.Error("ошибка добавления карты", zap.Error(err))
		return nil, grpcErr(err)
	}

	return &pb.CreateInfoResponse{
		Id:      card.ID,
		Version: version,
	}, nil
}

func (s *StorageServer) GetBankCardDetails(ctx context.Context, req *pb.GetInfoRequest) (*pb.StoredBankCardDetails, error) {
	userID, err := userIDFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	card, err := s.repo.Cards().Get(ctx, userID, req.GetId())
	if err != nil {
		s.lgr.Error("ошибка получения номера карты", zap.Error(err))
		return nil, grpcErr(err)
	}

	number, err := crypto.Decrypt(card.NumberEncrypted, s.key)
	if err != nil {
		s.lgr.Error("ошибка дешифрования номера карты", zap.Error(err))
		return nil, fmt.Errorf("ошибка дешифрования номера карты")
	}

	meta, err := unmarshalMeta(card.Meta)
	if err != nil {
		return nil, err
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
	userID, err := userIDFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	metaJSON, err := marshalMeta(req.Data.GetMeta())
	if err != nil {
		return nil, err
	}

	number, err := crypto.Encrypt([]byte(req.Data.GetNumber()), s.key)
	if err != nil {
		s.lgr.Error("ошибка шифрования номера карты", zap.Error(err))
		return nil, fmt.Errorf("ошибка шифрования номера карты %w", err)
	}

	last4n, err := parseLast4(req.Data.Number)
	if err != nil {
		return nil, err
	}

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

	newVersion, err := s.repo.Cards().Update(ctx, card, req.GetVersion())
	if err != nil {
		s.lgr.Error("ошибка обновления карты", zap.Error(err))
		return nil, grpcErr(err)
	}

	return &pb.UpdateInfoResponse{
		NewVersion: newVersion,
	}, nil
}

func (s *StorageServer) DeleteBankCardDetails(ctx context.Context, req *pb.DeleteInfoRequest) (*pb.DeleteInfoResponse, error) {
	userID, err := userIDFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	err = s.repo.Cards().Delete(ctx, userID, req.GetId(), req.GetVersion())
	if err != nil {
		s.lgr.Error("ошибка удаления карты", zap.Error(err))
		return nil, grpcErr(err)
	}

	return &pb.DeleteInfoResponse{}, nil
}

func (s *StorageServer) StartUpload(ctx context.Context, req *pb.StartUploadRequest) (*pb.StartUploadResponse, error) {
	userID, err := userIDFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	if req.GetFilename() == "" {
		return nil, status.Error(codes.InvalidArgument, "не указано имя файла")
	}

	uploadID, fileID, err := s.repo.Files().StartUpload(
		ctx,
		userID,
		req.GetFileId(),
		req.GetVersion(),
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
	userID, err := userIDFromCtx(stream.Context())
	if err != nil {
		return err
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
			return status.Error(codes.Internal, "ошибка отправки сообщения")
		}

		if req.GetUploadId() == "" {
			return status.Error(codes.InvalidArgument, "нужен upload_id")
		}
		if uploadID == "" {
			uploadID = req.GetUploadId()
		}
		if req.GetUploadId() != uploadID {
			return status.Error(codes.InvalidArgument, "все куски должны принадлежать одному upload_id")
		}
		if len(req.GetData()) == 0 {
			return status.Error(codes.InvalidArgument, "данные пустые")
		}

		if err := s.repo.Files().PutUploadChunk(
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
	userID, err := userIDFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	if req.GetUploadId() == "" {
		return nil, status.Error(codes.InvalidArgument, "нужен upload_id")
	}

	fileID, newVersion, err := s.repo.Files().CommitUpload(
		ctx,
		userID,
		req.GetUploadId(),
		req.GetFileId(),
		req.GetVersion(),
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
	userID, err := userIDFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	meta, err := s.repo.Files().GetFileMeta(ctx, userID, req.GetFileId())
	if err != nil {
		return nil, grpcErr(err)
	}

	outMeta, err := unmarshalMeta(meta.MetaJSON)
	if err != nil {
		return nil, err
	}

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
	userID, err := userIDFromCtx(stream.Context())
	if err != nil {
		return err
	}

	for chunk, err := range s.repo.Files().Chunks(stream.Context(), userID, req.GetFileId(), req.GetVersion()) {
		if err != nil {
			return grpcErr(err)
		}

		if err := stream.Send(&pb.DownloadFileChunk{
			ChunkNo: chunk.ChunkNo,
			Data:    chunk.Data,
		}); err != nil {
			return status.Errorf(codes.Internal, "отправка куска: %v", err)
		}
	}

	return nil
}

func (s *StorageServer) DeleteFile(ctx context.Context, req *pb.DeleteFileRequest) (*pb.DeleteInfoResponse, error) {
	userID, err := userIDFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	err = s.repo.Files().DeleteFile(ctx, userID, req.GetFileId(), req.GetVersion())
	if err != nil {
		return nil, grpcErr(err)
	}

	return &pb.DeleteInfoResponse{}, nil
}
