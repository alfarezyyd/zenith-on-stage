package storage

import (
	"fmt"
	"net/http"
	"zenith-on-stage/pkg/exception"
	"zenith-on-stage/pkg/helper"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	localStorageService *Manager
}

func NewHandler(localStorageService *Manager) *Handler {
	return &Handler{
		localStorageService: localStorageService,
	}
}

func (storageHandler *Handler) UploadFiles(ginContext *gin.Context) {
	err := ginContext.Request.ParseMultipartForm(512 << 20)
	fmt.Println(err)
	helper.CheckErrorOperation(err, exception.ParseGormError(err))
	multipartForm, err := ginContext.MultipartForm()
	fmt.Println(err)
	helper.CheckErrorOperation(err, exception.ParseGormError(err))
	files := multipartForm.File["files"]
	fmt.Println(files)
	var newTemporaryId []string
	for _, fileHeader := range files {
		tempUuid, err := uuid.NewV7()
		helper.CheckErrorOperation(err, exception.NewApplicationError(http.StatusInternalServerError, exception.ErrInternalServerError))
		tempUuidString := tempUuid.String()
		_, err = storageHandler.localStorageService.Disk().Put(fileHeader, fmt.Sprintf("/tmp/%s_%s", tempUuidString, fileHeader.Filename))
		helper.CheckErrorOperation(err, exception.NewApplicationError(http.StatusInternalServerError, exception.ErrInternalServerError))
		newTemporaryId = append(newTemporaryId, tempUuidString)
	}
	ginContext.JSON(http.StatusOK, helper.NewSuccessResponseWithEntries("Upload file success", newTemporaryId))

}
