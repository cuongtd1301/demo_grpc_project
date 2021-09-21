package service

import (
	"context"
	"demo-grpc/client/infrastructure"
	"demo-grpc/client/model"

	pb "demo-grpc/proto"
)

type mediaService struct {
}

type MediaService interface {
	DownloadImage(ctx context.Context, url string) (*model.FileUploadInfo, error)
}

func (s *mediaService) DownloadImage(ctx context.Context, url string) (*model.FileUploadInfo, error) {
	clientConn, err := infrastructure.GrpcClientConnect()
	if err != nil {
		return nil, err
	}
	c := pb.NewMediaServiceClient(clientConn)
	r, err := c.DownloadImage(ctx, &pb.ImageInfo{
		Url: url,
	})
	if err != nil {
		return nil, err
	}
	return &model.FileUploadInfo{
		Id:          int(r.GetId()),
		FileId:      r.GetFileId(),
		FileSize:    r.GetFileSize(),
		FileName:    r.GetFileName(),
		Ext:         r.GetExt(),
		MimeType:    r.GetMimeType(),
		CreatedTime: r.GetCreatedTime(),
		UpdateTime:  r.GetUpdatedTime(),
	}, nil
}

func NewMediaService() MediaService {
	return &mediaService{}
}
