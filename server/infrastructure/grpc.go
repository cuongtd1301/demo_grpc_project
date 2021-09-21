package infrastructure

const (
	grpcHost = "localhost"
	grpcPort = ":50001"
)

func GetGrpcAddress() string {
	return grpcHost + grpcPort
}
