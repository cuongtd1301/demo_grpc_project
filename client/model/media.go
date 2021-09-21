package model

type FileUploadInfo struct {
	Id          int    `json:"id"`
	FileId      int64  `json:"fileId"`
	FileSize    int64  `json:"fileSize"`
	FileName    string `json:"fileName"`
	Ext         string `json:"ext"`
	MimeType    string `json:"mimeType"`
	CreatedTime int64  `json:"createdTime"`
	UpdateTime  int64  `json:"updateTime"`
}

type ImageInfo struct {
	Url string `json:"url"`
}
