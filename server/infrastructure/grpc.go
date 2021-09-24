package infrastructure

import "fmt"

const (
// grpcHost = "localhost"
// grpcPort = ":50001"
)

func GetGrpcAddress() string {
	host := config.Grpc.Host
	port := config.Grpc.Port

	return fmt.Sprintf("%s:%s", host, port)
}
