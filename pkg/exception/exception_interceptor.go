package exception

import (
	"net/http"
	"runtime/debug"
	"zenith-on-stage/internal/model"
	"zenith-on-stage/pkg/logger"

	"github.com/gin-gonic/gin"
)

func Interceptor() gin.HandlerFunc {
	return func(ginContext *gin.Context) {
		defer func() {
			if occurredError := recover(); occurredError != nil {
				logger.Debug("panic occurred", occurredError)
				logger.Debug("stack trace:\n" + string(debug.Stack()))

				// Check if it's our custom error
				if clientError, ok := occurredError.(*ApplicationError); ok {
					ginContext.AbortWithStatusJSON(
						clientError.HttpStatusCode,
						&model.ResponseContract[any]{
							Success: false,
							Message: "",
							Entry:   nil,
							Entries: nil,
							Error: &model.ErrorDetail{
								Code:    clientError.ConventionStatusCode,
								Message: clientError.Message,
								Details: clientError.Details,
							},
						},
					)
					return
				}

				// Unknown error
				ginContext.AbortWithStatusJSON(
					http.StatusInternalServerError,
					model.NewResponseContractModel(false, "Internal server error", nil, nil),
				)
			}
		}()
		ginContext.Next()
	}
}
