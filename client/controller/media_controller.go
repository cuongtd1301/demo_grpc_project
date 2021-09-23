package controller

import (
	"demo-grpc/client/model"
	"demo-grpc/client/service"
	"io/ioutil"
	"net/http"

	"github.com/go-chi/render"
)

type mediaController struct {
	mediaService service.MediaService
}

type MediaController interface {
	ManageMedia(w http.ResponseWriter, r *http.Request)
	// DownloadImage(w http.ResponseWriter, r *http.Request)
}

func (c *mediaController) ManageMedia(w http.ResponseWriter, r *http.Request) {
	var res *model.Response

	var data model.MediaPayload
	query := r.URL.Query()
	data.Constructor = query.Get("constructor")
	data.Bucket = query.Get("bucket")
	data.Key = query.Get("key")

	r.ParseMultipartForm(25 << 20)
	file, header, err := r.FormFile("file")
	if err != nil {
		res := &model.Response{
			Data:    nil,
			Message: "Error retrieving the file:" + err.Error(),
			Success: false,
		}
		render.JSON(w, r, res)
		return
	}
	defer file.Close()

	byteFile, err := ioutil.ReadAll(file)
	if err != nil {
		res := &model.Response{
			Data:    nil,
			Message: "Error convert the file to bytes" + err.Error(),
			Success: false,
		}
		render.JSON(w, r, res)
		return
	}

	tmp, err := c.mediaService.ManageMedia(data, byteFile, header)
	if err != nil {
		res = &model.Response{
			Data:    nil,
			Message: err.Error(),
			Success: false,
		}
	} else {
		res = &model.Response{
			Data:    tmp,
			Message: "Media successfully.",
			Success: true,
		}
	}

	render.JSON(w, r, res)
}

func NewMediaController() MediaController {
	mediaService := service.NewMediaService()
	return &mediaController{
		mediaService: mediaService,
	}
}

// // Download image by url godoc
// // @tags media-manager-apis
// // @Summary Download image by url
// // @Description Download image by url
// // @Accept json
// // @Produce json
// // @Param ImageInfo body model.ImageInfo true "image information"
// // @Success 200 {object} model.Response
// // @Router /media/image [post]
// func (c *mediaController) DownloadImage(w http.ResponseWriter, r *http.Request) {
// 	var res *model.Response
// 	var data model.ImageInfo
// 	decoder := json.NewDecoder(r.Body)
// 	defer r.Body.Close()
// 	if err := decoder.Decode(&data); err != nil {
// 		w.WriteHeader(http.StatusBadRequest)
// 		http.Error(w, http.StatusText(400), 400)
// 		// infrastructure.ErrLog.Println(err)
// 		res = &model.Response{
// 			Data:    nil,
// 			Message: "Download Image failed: " + err.Error(),
// 			Success: false,
// 		}
// 		render.JSON(w, r, res)
// 		return
// 	}
// 	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
// 	defer cancel()
// 	tmp, err := c.mediaService.DownloadImage(ctx, data.Url)
// 	if err != nil {
// 		res = &model.Response{
// 			Data:    nil,
// 			Message: err.Error(),
// 			Success: false,
// 		}
// 	} else {
// 		res = &model.Response{
// 			Data:    tmp,
// 			Message: "Download Image successfully.",
// 			Success: true,
// 		}
// 	}
// 	render.JSON(w, r, res)
// }
