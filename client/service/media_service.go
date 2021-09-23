package service

import (
	"bytes"
	"context"
	"demo-grpc/client/infrastructure"
	"demo-grpc/client/model"
	"demo-grpc/client/utils"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"sort"
	"time"

	pb "demo-grpc/proto"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
)

const (
	PART_SIZE       = 5_242_880 // 5_242_880 minimim
	RETRIES         = 2
	LARGE_FILE_SIZE = 20_000_000
)

type mediaService struct {
}

type MediaService interface {
	ManageMedia(payload model.MediaPayload, byteData []byte, header *multipart.FileHeader) (*model.MediaResopnse, error)
	DownloadImage(ctx context.Context, url string) (*model.FileUploadInfo, error)
}

func (s *mediaService) ManageMedia(payload model.MediaPayload, byteData []byte, header *multipart.FileHeader) (*model.MediaResopnse, error) {
	clientConn, err := infrastructure.GrpcClientConnect()
	defer clientConn.Close()
	if err != nil {
		return nil, err
	}
	c := pb.NewMediaServiceClient(clientConn)
	switch payload.Constructor {
	case "upload":
		return UploadMedia(c, payload, byteData, header)
	case "download":
		return DownloadMedia(c, payload)
	default:
		return nil, errors.New("constructor not in (download, upload)")
	}
	// return nil, errors.New("something wrong")
}

func UploadMedia(c pb.MediaServiceClient, payload model.MediaPayload, byteData []byte, header *multipart.FileHeader) (*model.MediaResopnse, error) {
	if len(byteData) < LARGE_FILE_SIZE {
		return UploadNormalFile(c, payload, byteData, header)
	}
	return UploadLargeFile(c, payload, byteData, header)
}

func DownloadMedia(c pb.MediaServiceClient, payload model.MediaPayload) (*model.MediaResopnse, error) {
	contentLength, err := c.GetHeadFile(payload.Bucket, payload.Key)
	if err != nil {
		return nil, err
	}
	if contentLength < LARGE_FILE_SIZE {
		return DownloadNormalFile(c, payload)
	}
	return DownloadLargeFile(c, payload)
}

func DownloadNormalFile(c pb.MediaServiceClient, payload model.MediaPayload) (*model.MediaResopnse, error) {
	url := c.GetPresignedUrlDownloadFile(payload.Bucket, payload.Key)

	ctxRequest, cancelRequest := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancelRequest()
	req, err := http.NewRequestWithContext(ctxRequest, http.MethodGet, url, nil)
	if err != nil {
		fmt.Println("error creating request")
		return nil, err
	}
	// req.ContentLength = int64(len(byteData))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("failed making request:", err)
		return nil, err
	}
	defer resp.Body.Close()
	byteData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("failed reading")
	}
	log.Println(len(byteData))
	return &model.MediaResopnse{}, nil
}

func DownloadLargeFile(c pb.MediaServiceClient, payload model.MediaPayload) (*model.MediaResopnse, error) {
	listUrl := c.GetPresignedUrlDownloadPartFile(payload.Bucket, payload.Key, PART_SIZE)

	completedPartChannel := make(chan *DownloadPartResponse)
	defer close(completedPartChannel)

	for part := range listUrl {
		go downloadPartFileUrl(part, completedPartChannel)
	}
	partRes := []DownloadPartResponse{}
	for i := 0; i < len(listUrl); i++ {
		tmp := <-completedPartChannel
		if tmp == nil {
			return nil, errors.New("About download because some parts get error\n")
		}
		partRes = append(partRes, *tmp)
	}
	sort.Slice(partRes, func(i, j int) bool {
		return partRes[i].PartNum < partRes[j].PartNum
	})
	dataByte := []byte{}
	for i := range partRes {
		dataByte = append(dataByte, partRes[i].ByteData...)
	}

	dataByte

	return &model.MediaResopnse{}, nil
}

func downloadPartFileUrl(payload DownloadPartPayload, completedParts chan *DownloadPartResponse) {
	var try int
	for try <= RETRIES {
		ctxRequest, cancelRequest := context.WithTimeout(context.Background(), 1*time.Minute)
		defer cancelRequest()
		req, err1 := http.NewRequestWithContext(ctxRequest, http.MethodGet, payload.Url, nil)
		resp, err2 := http.DefaultClient.Do(req)
		defer resp.Body.Close()
		err := utils.FirstNonNil(err1, err2)
		// Download failed
		if err != nil {
			fmt.Println(err)
			// Max retries reached! Quitting
			if try == RETRIES {
				completedParts <- nil
				log.Println("partNum retries fail:", payload.PartNum)
				return
			} else {
				// Retrying
				try++
			}
		} else {
			// Download is done!
			byteData, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				fmt.Println("failed reading")
			}
			completedParts <- &DownloadPartResponse{
				ByteData: byteData,
				PartNum:  payload.PartNum,
			}
			fmt.Printf("Part %v complete\n", payload.PartNum)
			return
		}
	}
	return
}

func UploadNormalFile(c pb.MediaServiceClient, payload model.MediaPayload, byteData []byte, header *multipart.FileHeader) (*model.MediaResopnse, error) {
	url := c.GetPresignedUrlUploadFile(payload.Bucket, payload.Key)
	r := bytes.NewReader(byteData)
	ctxRequest, cancelRequest := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancelRequest()
	req, err := http.NewRequestWithContext(ctxRequest, http.MethodPut, url, r)
	if err != nil {
		fmt.Println("error creating request")
		return nil, err
	}
	req.ContentLength = int64(len(byteData))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("failed making request:", err)
		return nil, err
	}
	defer resp.Body.Close()
	return &model.MediaResopnse{}, nil
}

func UploadLargeFile(c pb.MediaServiceClient, payload model.MediaPayload, byteData []byte, header *multipart.FileHeader) (*model.MediaResopnse, error) {
	// ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	// defer cancel()
	uploadId, listPresignedUrlPart, err := c.GetPresignedUrlUploadLargeFile(payload.Bucket, payload.Key, len(byteData), PART_SIZE)
	if err != nil {
		return nil, err
	}
	// multipart upload
	var start, currentSize int
	var remaining = len(byteData)
	var partNum = 1
	completedPartChannel := make(chan *PresignedUrlPart)
	defer close(completedPartChannel)
	for start = 0; remaining != 0; start += PART_SIZE {
		if remaining < PART_SIZE*2 {
			currentSize = remaining
		} else {
			currentSize = PART_SIZE
		}
		go uploadPartFileUsingPresignedUrl(listPresignedUrlPart[partNum-1], byteData[start:start+currentSize], completedPartChannel)
		// Detract the current part size from remaining
		remaining -= currentSize

		partNum++
	}

	// append completedPart
	completedAllPart := true
	listCompletedParts := []*s3.CompletedPart{}
	listInfoPart := []PresignedUrlPart{}
	for i := 0; i < partNum-1; i++ {
		tmp := <-completedPartChannel
		if tmp == nil || !tmp.Success {
			log.Printf("About upload because some parts get error\n")
			completedAllPart = false
			// return errors.New("About upload because some parts get error\n")
		}
		listInfoPart = append(listInfoPart, *tmp)
		listCompletedParts = append(listCompletedParts, &s3.CompletedPart{
			ETag:       aws.String(tmp.ETag),
			PartNumber: aws.Int64(int64(tmp.PartNumber)),
		})
	}
	// // Import to redis
	// value, _ := json.Marshal(listInfoPart)
	// client := infrastructure.GetRedisClient()
	// ctxTimeout, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	// defer cancel()
	// err = client.HSet(ctxTimeout, infrastructure.GetBucketName(), stats.Name(), string(value)).Err()
	// if err != nil {
	// 	return
	// }
	// complete multipart upload
	if !completedAllPart {
		c.AbortMultipartUpload(payload.Bucket, payload.Key, uploadId)
		err = errors.New("About upload because some parts get error\n")
		return nil, err
	}
	_, err = c.CompleteMultipartUpload(payload.Bucket, payload.Key, uploadId, listCompletedParts)
	if err != nil {
		return nil, err
	}
	return &model.MediaResopnse{}, nil
}

func uploadPartFileUsingPresignedUrl(part PresignedUrlPart, fileBytes []byte, completedParts chan *PresignedUrlPart) {
	var try int
	for try <= RETRIES {
		body := bytes.NewReader(fileBytes)
		req, err1 := http.NewRequest(http.MethodPut, part.Url, body)
		ctxTimeout, cancel := context.WithTimeout(context.Background(), time.Second*60)
		defer cancel()
		resp, err2 := http.DefaultClient.Do(req.WithContext(ctxTimeout))
		err := utils.FirstNonNil(err1, err2)
		// Upload failed
		if err != nil {
			fmt.Println(err)
			// Max retries reached! Quitting
			if try == RETRIES {
				completedParts <- &PresignedUrlPart{
					UploadId:   part.UploadId,
					ETag:       resp.Header.Get("Etag"),
					PartNumber: part.PartNumber,
					Url:        part.Url,
					Success:    false,
				}
				return
			} else {
				// Retrying
				try++
			}
		} else {
			// Upload is done!
			completedParts <- &PresignedUrlPart{
				UploadId:   part.UploadId,
				ETag:       resp.Header.Get("Etag"),
				PartNumber: part.PartNumber,
				Url:        part.Url,
				Success:    true,
			}
			fmt.Printf("Part %v complete\n", part.PartNumber)
			return
		}
	}
	return
}

func (s *mediaService) DownloadImage(ctx context.Context, url string) (*model.FileUploadInfo, error) {
	clientConn, err := infrastructure.GrpcClientConnect()
	defer clientConn.Close()
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

type PresignedUrlPart struct {
	Key        string
	UploadId   string
	ETag       string
	PartNumber int
	Url        string
	Success    bool
}

type DownloadPartPayload struct {
	Url     string
	PartNum int
}

type DownloadPartResponse struct {
	ByteData []byte
	PartNum  int
}
