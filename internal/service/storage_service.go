package service

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/konkovaanna23/gophkeeper/internal/auth"
	"github.com/konkovaanna23/gophkeeper/internal/repository"
	pb "github.com/konkovaanna23/gophkeeper/pkg/keeperservice"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type StorageService struct {
	pb.UnimplementedAuthServiceServer
	repo repository.StoreRepository
}

func NewStorageService(repo repository.StoreRepository) *StorageService {
	return &StorageService{
		repo: repo,
	}
}

func (s *StorageService) SetAuthInfo(ctx context.Context, req *pb.AuthInfo) (*emptypb.Empty, error) {
	userID, ok := ctx.Value(auth.UserIDKey).(string)
	if !ok || userID == "" {
		return nil, status.Error(codes.Unauthenticated, "user is not authenticated")
	}

	auth, err := ConvertAuthDataPBToModel(req)
	if err != nil {
		return nil, fmt.Errorf("error convert AuthInfo %s", err.Error())
	}
	err = s.repo.SaveAuthData(ctx, userID, auth)
	if err != nil {
		return nil, fmt.Errorf("error save data: %s", err.Error())
	}

	return &emptypb.Empty{}, nil
}

func (s *StorageService) SetTextInfo(ctx context.Context, req *pb.TextInfo) (*emptypb.Empty, error) {
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

func (s *StorageService) SetFileChunkInfo(stream pb.StorageService_SetFileChunkInfoServer) error {

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

func (s *StorageService) SetBankCardDetails(ctx context.Context, req *pb.BankCardDetails) (*emptypb.Empty, error) {
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

func (s *StorageService) GetAuthInfo(ctx context.Context, req *pb.RequestSite) (*pb.AuthInfoList, error) {
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
}

func (s *StorageService) GetTextInfo(ctx context.Context, req *pb.RequestTitle) (*pb.TextInfoList, error) {
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

func (s *StorageService) GetFileChunkInfo(req *pb.RequestFileName, stream pb.StorageService_GetFileChunkInfoServer) error {
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

func (s *StorageService) GetBankCardDetails(ctx context.Context, req *pb.RequestCardNumber) (*pb.BankCardDetailsList, error) {
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
