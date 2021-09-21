package service

import (
	"context"
	"demo-grpc/server/infrastructure"
	"demo-grpc/server/model"
	"demo-grpc/server/utils"
	"log"
	"time"
)

func Insert(info model.FileUploadInfo) (*model.FileUploadInfo, error) {
	db := infrastructure.GetDB()

	now := time.Now()
	info.CreatedAt = &now
	info.UpdatedAt = &now
	info.CreatedTime = now.Unix()
	info.UpdateTime = now.Unix()
	info.FileId = utils.GenUuid()

	ctxTimeout, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()
	result, err := db.NamedExecContext(ctxTimeout, `INSERT INTO file_upload_infos (file_id, file_size, file_name, ext, mime_type, created_time, updated_time, created_at, updated_at) 
		VALUES (:file_id, :file_size, :file_name, :ext, :mime_type, :created_time, :updated_time, :created_at, :updated_at)`, &info)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	var tmp model.FileUploadInfo
	id, err := result.LastInsertId()
	err = db.Get(&tmp, "SELECT * FROM file_upload_infos where id = ?;", id)
	return &tmp, err
}
