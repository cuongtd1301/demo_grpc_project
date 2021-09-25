package service

import (
	"context"
	pb "demo-grpc/proto"
	"demo-grpc/server/infrastructure"
	"errors"
	"fmt"
	"io/ioutil"
	"sort"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
)

type server struct {
	pb.UnimplementedMediaServiceServer
}

func (s *server) GetHeadFile(ctx context.Context, in *pb.FileInput) (*pb.FileHeader, error) {
	svc := s3.New(infrastructure.GetAwsSession())
	out, err := svc.HeadObjectWithContext(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(in.GetBucket()),
		Key:    aws.String(in.GetKey()),
	})
	if err != nil {
		fmt.Println("Failed to HeadObjectWithContext: ", err)
		return nil, err
	}
	return &pb.FileHeader{
		ContentLength: *out.ContentLength,
	}, nil
}

func (s *server) GetObjectByRange(ctx context.Context, in *pb.FilePartInput) (*pb.FilePartObject, error) {
	svc := s3.New(infrastructure.GetAwsSession())
	result, err := svc.GetObjectWithContext(ctx, &s3.GetObjectInput{
		Bucket: aws.String(in.GetBucket()),
		Key:    aws.String(in.GetKey()),
		Range:  aws.String(fmt.Sprintf("bytes=%d-%d", in.GetRangeStart(), in.GetRangeEnd())),
	})
	if err != nil {
		return nil, err
	}
	tmpData, _ := ioutil.ReadAll(result.Body)
	return &pb.FilePartObject{
		Data: tmpData,
	}, nil
}

func (s *server) GetPresignedUrlDownloadFile(ctx context.Context, in *pb.FileInput) (*pb.FileDownloadUrl, error) {
	svc := s3.New(infrastructure.GetAwsSession())
	res, _ := svc.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(in.GetBucket()),
		Key:    aws.String(in.GetKey()),
	})
	url, err := res.Presign(15 * time.Minute)
	if err != nil {
		fmt.Println("Failed to generate a pre-signed url: ", err)
		return nil, err
	}
	return &pb.FileDownloadUrl{
		Url: url,
	}, nil
}

func (s *server) GetPresignedUrlDownloadPartFile(ctx context.Context, in *pb.LargeFileInput) (*pb.LargeFileResponse, error) {
	largeFileResponse := &pb.LargeFileResponse{}
	bucket := in.GetBucket()
	key := in.GetKey()
	contentLength := in.GetContentLength()
	partSize := in.GetPartSize()
	svc := s3.New(infrastructure.GetAwsSession())
	partNum := int64(1)
	for startRange := int64(0); startRange < contentLength; startRange += partSize {
		var try int
		for try <= RETRIES {
			res, _ := svc.GetObjectRequest(&s3.GetObjectInput{
				Bucket: aws.String(bucket),
				Key:    aws.String(key),
				Range:  aws.String(fmt.Sprintf("bytes=%d-%d", startRange, startRange+partSize-1)),
			})
			url, err := res.Presign(15 * time.Minute)
			if err != nil {
				// Max retries reached! Quitting
				if try == RETRIES {
					return nil, err
				} else {
					// Retrying
					try++
					continue
				}
			}
			largeFileResponse.PartFile = append(largeFileResponse.PartFile, &pb.LargeFileResponse_PartFile{
				Url:        url,
				PartNumber: partNum,
			})
			partNum++
			break
		}
	}
	return largeFileResponse, nil
}

func (s *server) GetPresignedUrlUploadFile(ctx context.Context, in *pb.FileInput) (*pb.FileUploadUrl, error) {
	svc := s3.New(infrastructure.GetAwsSession())
	res, _ := svc.PutObjectRequest(&s3.PutObjectInput{
		Bucket: aws.String(in.GetBucket()),
		Key:    aws.String(in.GetKey()),
	})
	url, err := res.Presign(15 * time.Minute)
	if err != nil {
		fmt.Println("Failed to generate a pre-signed url: ", err)
		return nil, err
	}
	return &pb.FileUploadUrl{
		Url: url,
	}, nil
}

func (s *server) GetPresignedUrlUploadLargeFile(ctx context.Context, in *pb.LargeFileInput) (*pb.LargeFileResponse, error) {
	largeFileResponse := &pb.LargeFileResponse{}
	svc := s3.New(infrastructure.GetAwsSession())
	out, err := svc.CreateMultipartUploadWithContext(ctx, &s3.CreateMultipartUploadInput{
		Bucket: aws.String(in.GetBucket()),
		Key:    aws.String(in.GetKey()),
	})
	if err != nil {
		fmt.Println("Failed to CreateMultipartUpload:", err)
		return nil, err
	}
	largeFileResponse.UploadId = *out.UploadId
	// Get part url
	var start, currentSize int64
	var remaining = in.GetContentLength()
	var partNum int64 = 1
	partSize := in.GetPartSize()
	for start = 0; remaining != 0; start += partSize {
		if remaining < partSize*2 {
			currentSize = remaining
		} else {
			currentSize = partSize
		}
		var try int
		for try <= RETRIES {
			req, _ := svc.UploadPartRequest(&s3.UploadPartInput{
				Bucket:        out.Bucket,
				Key:           out.Key,
				PartNumber:    aws.Int64(int64(partNum)),
				UploadId:      out.UploadId,
				ContentLength: aws.Int64(int64(currentSize)),
			})

			url, err := req.Presign(15 * time.Minute)
			if err != nil {
				if try == RETRIES {
					// Max retries reached! Quitting
					return nil, errors.New(fmt.Sprintln("Failed to generate a pre-signed url of part ", partNum, ":", err))
				} else {
					// Retrying
					try++
					continue
				}
			}
			largeFileResponse.PartFile = append(largeFileResponse.PartFile, &pb.LargeFileResponse_PartFile{
				Url:        url,
				PartNumber: partNum,
			})
			// Detract the current part size from remaining
			remaining -= currentSize
			partNum++
			break
		}
	}
	return largeFileResponse, nil
}

func (s *server) CompleteMultipartUpload(ctx context.Context, in *pb.UploadCompleteInput) (*pb.UploadCompleteResponse, error) {
	listCompletedParts := []*s3.CompletedPart{}
	inputCompletedParts := in.GetCompletedPart()
	for i := range inputCompletedParts {
		listCompletedParts = append(listCompletedParts, &s3.CompletedPart{
			ETag:       aws.String(inputCompletedParts[i].GetETag()),
			PartNumber: aws.Int64(inputCompletedParts[i].GetPartNumber()),
		})
	}
	sort.Slice(listCompletedParts, func(i, j int) bool {
		return int(*listCompletedParts[i].PartNumber) < int(*listCompletedParts[j].PartNumber)
	})
	svc := s3.New(infrastructure.GetAwsSession())
	_, err := svc.CompleteMultipartUploadWithContext(ctx, &s3.CompleteMultipartUploadInput{
		Bucket:   aws.String(in.GetBucket()),
		Key:      aws.String(in.GetKey()),
		UploadId: aws.String(in.GetUploadId()),
		MultipartUpload: &s3.CompletedMultipartUpload{
			Parts: listCompletedParts,
		},
	})
	if err != nil {
		fmt.Println("Failed to AbortMultipartUploadInput: ", err)
		return nil, err
	}
	return &pb.UploadCompleteResponse{
		Status: "200",
	}, nil
}

func (s *server) AbortMultipartUpload(ctx context.Context, in *pb.UploadCompleteInput) (*pb.UploadCompleteResponse, error) {
	svc := s3.New(infrastructure.GetAwsSession())
	_, err := svc.AbortMultipartUploadWithContext(ctx, &s3.AbortMultipartUploadInput{
		Bucket:   aws.String(in.GetBucket()),
		Key:      aws.String(in.GetKey()),
		UploadId: aws.String(in.GetUploadId()),
	})
	if err != nil {
		fmt.Println("Failed to AbortMultipartUploadInput: ", err)
		return nil, err
	}
	return &pb.UploadCompleteResponse{
		Status: "200",
	}, nil
}

func GetServerGrpcStruct() *server {
	return &server{}
}

// func (s *server) DownloadImage(ctx context.Context, in *pb.ImageInfo) (*pb.FileUploadInfo, error) {
// 	url := in.GetUrl()
// 	//Get the response bytes from the url
// 	response, err := http.Get(url)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer response.Body.Close()
// 	if response.StatusCode != 200 {
// 		return nil, errors.New("Received non 200 response code")
// 	}
// 	if !strings.Contains(response.Header.Get("Content-Type"), "image") {
// 		return nil, errors.New("Content-Type not image")
// 	}
// 	// ioutil.WriteFile()
// 	file, err := ioutil.TempFile(infrastructure.GetTempPath(), "image*.jpg")
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer os.Remove(file.Name())
// 	//Write the bytes to the fiel
// 	b, err := ioutil.ReadAll(response.Body)
// 	if err != nil {
// 		return nil, err
// 	}
// 	err = ioutil.WriteFile(file.Name(), b, 0644)
// 	if err != nil {
// 		return nil, err
// 	}
// 	info, err := UploadFileToBucketV2(file, response.Header.Get("Content-Type"))
// 	if err != nil {
// 		return nil, err
// 	}
// 	return &pb.FileUploadInfo{
// 		Id:          int32(info.Id),
// 		FileId:      info.FileId,
// 		FileSize:    info.FileSize,
// 		FileName:    info.FileName,
// 		Ext:         info.Ext,
// 		MimeType:    info.MimeType,
// 		CreatedTime: info.CreatedTime,
// 		UpdatedTime: info.UpdateTime,
// 	}, nil
// }
