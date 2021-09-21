package infrastructure

import (
	"context"
	"time"

	"google.golang.org/grpc"
)

const (
	grpcHost = "localhost"
	grpcPort = ":50001"
)

var (
	clientConnect *grpc.ClientConn
)

func GrpcClientConnect() (*grpc.ClientConn, error) {
	address := grpcHost + grpcPort
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	clientConn, err := grpc.DialContext(ctx, address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return nil, err
	}
	return clientConn, nil
}

func InitGrpc() error {
	var err error
	address := grpcHost + grpcPort
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	clientConnect, err = grpc.DialContext(ctx, address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return err
	}
	return nil
	// defer conn.Close()
}

func GetClientConnect() *grpc.ClientConn {
	return clientConnect
}
