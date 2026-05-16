package storage

import "github.com/gin-gonic/gin"

type Controller interface {
	UploadFiles(ginContext *gin.Context)
}
