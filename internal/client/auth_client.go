package client

import (
	"context"

	pb "github.com/konkovaanna23/gophkeeper/pkg/keeperservice"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type AuthServiceClient struct {
	conn   *grpc.ClientConn
	client pb.AuthServiceClient
}

func NewAuthServiceClient(grpcAddr string) (*AuthServiceClient, error) {
	conn, err := grpc.NewClient(
		grpcAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}

	return &AuthServiceClient{
		conn:   conn,
		client: pb.NewAuthServiceClient(conn),
	}, nil
}

func (c *AuthServiceClient) Close() error {
	if c == nil || c.conn == nil {
		return nil
	}
	return c.conn.Close()
}

func (c *AuthServiceClient) Register(ctx context.Context, login, password string) (string, error) {
	resp, err := c.client.Register(ctx, &pb.RegisterRequest{
		Login:    login,
		Password: password,
	})
	if err != nil {
		return "", err
	}
	return resp.UserId, nil
}

func (c *AuthServiceClient) Login(ctx context.Context, login, password string) (string, error) {
	resp, err := c.client.Login(ctx, &pb.LoginRequest{
		Login:    login,
		Password: password})
	if err != nil {
		return "", err
	}
	return resp.AccessToken, nil
}
