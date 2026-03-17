package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/konkovaanna23/gophkeeper/internal/auth"
	"github.com/konkovaanna23/gophkeeper/internal/crypto"
	"github.com/konkovaanna23/gophkeeper/internal/model"
	"github.com/konkovaanna23/gophkeeper/internal/repository"
	pb "github.com/konkovaanna23/gophkeeper/pkg/keeperservice"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
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

func (s *StorageServer) CreateAuthInfo(ctx context.Context, req *pb.CreateAuthInfoRequest) (*pb.CreateAuthInfoResponse, error) {
	userID, ok := ctx.Value(auth.UserIDKey).(string)
	if !ok || userID == "" {
		return nil, status.Error(codes.Unauthenticated, "user is not authenticated")
	}

	metaJSON, _ := json.Marshal(req.GetData().GetMeta())
	password, err := crypto.Encrypt([]byte(req.GetData().GetPassword()), []byte("key"))
	if err != nil {
		return nil, fmt.Errorf("error crypto password %w", err)
	}
	auth := &model.AuthData{
		ID:       NewUUID(),
		UserID:   userID,
		Login:    req.GetData().Login,
		Password: password,
		Meta:     metaJSON,
	}

	version, err := s.repo.CreateAuthData(ctx, auth)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to create item")
	}

	return &pb.CreateAuthInfoResponse{
		Id:      auth.ID,
		Version: version,
	}, nil
}

func (s *StorageServer) GetAuthInfo(ctx context.Context, req *pb.GetAuthInfoRequest) (*pb.GetAuthInfoResponse, error) {
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
	item := &pb.StoredAuthInfo{
		Id:        auth.ID,
		Data:      data,
		UpdatedAt: timestamppb.New(auth.UpdatedAt),
		Version:   auth.Version,
	}

	return &pb.GetAuthInfoResponse{
		Item: item,
	}, nil
}

func (s *StorageServer) UpdateAuthInfo(ctx context.Context, req *pb.UpdateAuthInfoRequest) (*pb.UpdateAuthInfoResponse, error) {
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

	return &pb.UpdateAuthInfoResponse{
		NewVersion: newVersion,
	}, nil
}

func (s *StorageServer) DeleteAuthInfo(ctx context.Context, req *pb.DeleteAuthInfoRequest) (*pb.DeleteAuthInfoResponse, error) {
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

	return &pb.DeleteAuthInfoResponse{}, nil
}

func (s *StorageServer) SetTextInfo(ctx context.Context, req *pb.TextInfo) (*emptypb.Empty, error) {
	userID, ok := ctx.Value(auth.UserIDKey).(string)
	if !ok || userID == "" {
		return nil, status.Error(codes.Unauthenticated, "user is not authenticated")
	}

	text := ConvertTextDataPBToModel(req)
	err := s.repo.SaveTextData(ctx, userID, text)
	if err != nil {
		return nil, fmt.Errorf("error save data: %s", err.Error())
	}

	return &emptypb.Empty{}, nil
}

func (s *StorageServer) SetFileChunkInfo(stream pb.StorageService_SetFileChunkInfoServer) error {

	userID, ok := stream.Context().Value(auth.UserIDKey).(string)
	if !ok || userID == "" {
		return status.Error(codes.Unauthenticated, "user is not authenticated")
	}

	chunkNum := 1

	for {
		in, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				return stream.SendAndClose(&emptypb.Empty{})
			}
			return err
		}

		file := ConvertFileChunkPBToModel(in, chunkNum)

		err = s.repo.SaveFileChunk(stream.Context(), userID, file)
		if err != nil {
			return fmt.Errorf("failed to save chunk %d: %w", chunkNum, err)
		}

		chunkNum++
	}
}

func (s *StorageServer) SetBankCardDetails(ctx context.Context, req *pb.BankCardDetails) (*emptypb.Empty, error) {
	userID, ok := ctx.Value(auth.UserIDKey).(string)
	if !ok || userID == "" {
		return nil, status.Error(codes.Unauthenticated, "user is not authenticated")
	}

	card := ConvertBankCardPBToModel(req)
	err := s.repo.SaveBankCardData(ctx, userID, card)
	if err != nil {
		return nil, fmt.Errorf("error save data: %s", err.Error())
	}

	return &emptypb.Empty{}, nil
}

/*func (s *StorageServer) GetAuthInfo(ctx context.Context, req *pb.RequestSite) (*pb.AuthInfoList, error) {
	userID, ok := ctx.Value(auth.UserIDKey).(string)
	if !ok || userID == "" {
		return nil, status.Error(codes.Unauthenticated, "user is not authenticated")
	}

	data, err := s.repo.GetAuthData(ctx, userID, req.GetSite())
	if err != nil {
		return nil, fmt.Errorf("failed to get data from db %w", err)
	}
	list := make([]*pb.AuthInfo, len(data))
	for i, d := range data {
		list[i], _ = ConvertAuthDataModelToPB(d)
	}
	return &pb.AuthInfoList{List: list}, nil
}*/

func (s *StorageServer) GetTextInfo(ctx context.Context, req *pb.RequestTitle) (*pb.TextInfoList, error) {
	userID, ok := ctx.Value(auth.UserIDKey).(string)
	if !ok || userID == "" {
		return nil, status.Error(codes.Unauthenticated, "user is not authenticated")
	}

	data, err := s.repo.GetTextData(ctx, userID, req.GetTitle())
	if err != nil {
		return nil, fmt.Errorf("failed to get data from db %w", err)
	}
	list := make([]*pb.TextInfo, len(data))
	for i, d := range data {
		list[i] = ConvertTextDataModelToPB(d)
	}
	return &pb.TextInfoList{List: list}, nil

}

func (s *StorageServer) GetFileChunkInfo(req *pb.RequestFileName, stream pb.StorageService_GetFileChunkInfoServer) error {
	userID, ok := stream.Context().Value(auth.UserIDKey).(string)
	if !ok || userID == "" {
		return status.Error(codes.Unauthenticated, "user is not authenticated")
	}
	chunkNum := 1
	for {
		file, err := s.repo.GetFileChunk(stream.Context(), userID, req.FileName, chunkNum)
		if err != nil {
			if errors.Is(err, repository.ErrorNotContent) {
				return nil
			} else {
				return fmt.Errorf("failed to get data from db %w", err)
			}
		}
		if err := stream.Send(ConvertFileChunkModelToPB(file)); err != nil {
			return fmt.Errorf("failed to send chunk %d: %w", chunkNum, err)
		}
		chunkNum++
	}
}

func (s *StorageServer) GetBankCardDetails(ctx context.Context, req *pb.RequestCardNumber) (*pb.BankCardDetailsList, error) {
	userID, ok := ctx.Value(auth.UserIDKey).(string)
	if !ok || userID == "" {
		return nil, status.Error(codes.Unauthenticated, "user is not authenticated")
	}

	data, err := s.repo.GetBankCardData(ctx, userID, req.GetLast4())
	if err != nil {
		return nil, fmt.Errorf("failed to get data from db %w", err)
	}
	list := make([]*pb.BankCardDetails, len(data))
	for i, d := range data {
		list[i] = ConvertBankCardModelToPB(d)
	}
	return &pb.BankCardDetailsList{List: list}, nil
}
