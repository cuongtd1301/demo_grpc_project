package infrastructure

import (
	"log"
	"os"
	"path"

	"github.com/BurntSushi/toml"
)

var (
	config   Config
	rootPath string
	tempPath string
)

func loadEnvParameters() {
	root, _ := os.Getwd()
	rootPath = path.Join(root, "server")
	tempPath = path.Join(rootPath, "temp")

	_, err := toml.DecodeFile(path.Join(rootPath, "infrastructure", "config.toml"), &config)
	if err != nil {
		log.Fatalln(err)
		return
	}
}

func init() {
	loadEnvParameters()
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

type Config struct {
	Databases Databases `toml:"databases"`
	Aws       Aws       `toml:"aws"`
	Grpc      Grpc      `toml:"grpc"`
	Redis     Redis     `toml:"redis"`
}

type Databases struct {
	Username string `toml:"username"`
	Password string `toml:"password"`
	Protocol string `toml:"protocol"`
	Ip       string `toml:"ip"`
	DbPort   string `toml:"dbPort"`
	DbName   string `toml:"dbName"`
}
type Aws struct {
	Region string `toml:"region"`
	Bucket string `toml:"bucket"`
}

type Grpc struct {
	Host string `toml:"host"`
	Port string `toml:"port"`
}

type Redis struct {
	Host     string `toml:"host"`
	Port     string `toml:"port"`
	Password string `toml:"password"`
	Db       int    `toml:"db"`
}
