package infrastructure

import (
	"os"
)

var (
	appPort     string
	rootPath    string
	httpSwagger string

	// dbHost     string
	// dbPort     string
	// dbUser     string
	// dbPassword string
	// dbName     string
)

func loadEnvParameters() {
	rootPath, _ = os.Getwd()
	appPort = "15001"

	httpSwagger = "http://localhost:15001/api/v1/swagger/doc.json"
}

func init() {
	loadEnvParameters()

	// if err := InitGrpc(); err != nil {
	// 	log.Fatalf("did not connect: %v", err)
	// }
}

// GetHTTPSwagger export link swagger
func GetHTTPSwagger() string {
	return httpSwagger
}

// GetAppPort export app port
func GetAppPort() string {
	return appPort
}

// GetRootPath export root path system
func GetRootPath() string {
	return rootPath
}
