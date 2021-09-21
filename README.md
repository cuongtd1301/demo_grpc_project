## Init go module before run
go mod init demo-grpc

## Run protoc
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/media.proto

## Command run server
go run server/main.go

## Generate swagger client side
swag init -d client -o client/docs

## Command run client
go run client/main.go