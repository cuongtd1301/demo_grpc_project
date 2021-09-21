package main

import (
	"demo-grpc/client/infrastructure"
	"demo-grpc/client/router"
	"log"
	"net/http"
)

// @title Swagger demo grpc
// @verson 1.0
// @description This is list apis in project

// @host localhost:15001
// @BasePath /api/v1

func main() {
	log.Println(infrastructure.GetHTTPSwagger())
	log.Fatal(http.ListenAndServe(":"+infrastructure.GetAppPort(), router.Router()))
}
