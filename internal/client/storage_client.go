package client

import (
	pb "github.com/konkovaanna23/gophkeeper/pkg/keeperservice"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

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
