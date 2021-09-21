package controller

import (
	"context"
	"demo-grpc/client/model"
	"demo-grpc/client/service"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/render"
)

type mediaController struct {
	mediaService service.MediaService
}

type MediaController interface {
	DownloadImage(w http.ResponseWriter, r *http.Request)
}

// Download image by url godoc
// @tags media-manager-apis
// @Summary Download image by url
// @Description Download image by url
// @Accept json
// @Produce json
// @Param ImageInfo body model.ImageInfo true "image information"
// @Success 200 {object} model.Response
// @Router /media/image [post]
func (c *mediaController) DownloadImage(w http.ResponseWriter, r *http.Request) {
	var res *model.Response

	var data model.ImageInfo
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()
	if err := decoder.Decode(&data); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		http.Error(w, http.StatusText(400), 400)
		// infrastructure.ErrLog.Println(err)
		res = &model.Response{
			Data:    nil,
			Message: "Download Image failed: " + err.Error(),
			Success: false,
		}
		render.JSON(w, r, res)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	tmp, err := c.mediaService.DownloadImage(ctx, data.Url)
	if err != nil {
		res = &model.Response{
			Data:    nil,
			Message: err.Error(),
			Success: false,
		}
	} else {
		res = &model.Response{
			Data:    tmp,
			Message: "Download Image successfully.",
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
