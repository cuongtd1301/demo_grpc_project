package service

import (
	"context"
	pb "demo-grpc/proto"
	"demo-grpc/server/infrastructure"
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

type server struct {
	pb.UnimplementedMediaServiceServer
}

func (s *server) DownloadImage(ctx context.Context, in *pb.ImageInfo) (*pb.FileUploadInfo, error) {
	url := in.GetUrl()
	//Get the response bytes from the url
	response, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		return nil, errors.New("Received non 200 response code")
	}
	if !strings.Contains(response.Header.Get("Content-Type"), "image") {
		return nil, errors.New("Content-Type not image")
	}
	// ioutil.WriteFile()
	file, err := ioutil.TempFile(infrastructure.GetTempPath(), "image*.jpg")
	if err != nil {
		return nil, err
	}
	defer os.Remove(file.Name())
	//Write the bytes to the fiel
	b, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	err = ioutil.WriteFile(file.Name(), b, 0644)
	if err != nil {
		return nil, err
	}

	info, err := UploadFileToBucketV2(file, response.Header.Get("Content-Type"))
	if err != nil {
		return nil, err
	}
	return &pb.FileUploadInfo{
		Id:          int32(info.Id),
		FileId:      info.FileId,
		FileSize:    info.FileSize,
		FileName:    info.FileName,
		Ext:         info.Ext,
		MimeType:    info.MimeType,
		CreatedTime: info.CreatedTime,
		UpdatedTime: info.UpdateTime,
	}, nil
}

func GetServerGrpcStruct() *server {
	return &server{}
}
