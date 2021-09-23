package infrastructure

import (
	"log"
	"os"
	"path"
)

var (
	rootPath string
	tempPath string
)

func loadEnvParameters() {
	root, _ := os.Getwd()
	rootPath = path.Join(root, "server")
	tempPath = path.Join(rootPath, "temp")
}

func init() {
	loadEnvParameters()
	// loadDriveService()
	loadAwsService()
	loadDatabase()
	loadRedis()
	log.Println("Load parameters succussful!")
}

// GetRootPath export root path server
func GetRootPath() string {
	return rootPath
}

// GetTempPath export temp path server
func GetTempPath() string {
	return tempPath
}
