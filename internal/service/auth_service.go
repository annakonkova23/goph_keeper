package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/konkovaanna23/gophkeeper/internal/auth"
	"github.com/konkovaanna23/gophkeeper/internal/repository"
	pb "github.com/konkovaanna23/gophkeeper/pkg/keeperservice"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthServer struct {
	pb.UnimplementedAuthServiceServer
	jwtManager *auth.JWTManager
	repo       repository.StoreRepository
}

func NewAuthServer(repo repository.StoreRepository, jwtManager *auth.JWTManager) *AuthServer {
	return &AuthServer{
		jwtManager: jwtManager,
		repo:       repo,
	}
}

func (s *AuthServer) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	if req.GetLogin() == "" || req.GetPassword() == "" {
		return nil, status.Error(codes.InvalidArgument, "email and password are required")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.GetPassword()), bcrypt.DefaultCost)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to hash password")
	}

	userID := NewUUID()

	err = s.repo.CreateUser(ctx, userID, req.GetLogin(), string(hash))
	if err != nil {
		if errors.Is(err, repository.ErrorConflict) {
			return nil, status.Error(codes.AlreadyExists, "user already exists")
		}
		return nil, status.Error(codes.Internal, "failed to create user")
	}

	return &pb.RegisterResponse{UserId: userID}, nil
}

func (s *AuthServer) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	if req.GetLogin() == "" || req.GetPassword() == "" {
		return nil, status.Error(codes.InvalidArgument, "email and password are required")
	}

	user, passwordHash, err := s.repo.GetUserByLogin(ctx, req.GetLogin())
	if err != nil {
		if errors.Is(err, repository.ErrorUserNotFound) {
			return nil, status.Error(codes.Unauthenticated, "invalid email or password")
		}
		return nil, status.Error(codes.Internal, "failed to get user")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.GetPassword())); err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid email or password")
	}

	token, err := s.jwtManager.Generate(user)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to generate token")
	}

	return &pb.LoginResponse{AccessToken: token}, nil
}

func NewUUID() string {
	return uuid.NewString()
}
