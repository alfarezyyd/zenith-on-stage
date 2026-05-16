package middleware

import (
	"net/http"
	"strings"
	"zenith-on-stage/pkg/exception"
	"zenith-on-stage/pkg/helper"

	"github.com/gin-gonic/gin"
)

func HasPermission(requiredPermissions ...string) gin.HandlerFunc {
	return func(ginContext *gin.Context) {
		// Ambil permission user dari context (diparsing saat login)
		userJwtClaim := helper.ExtractJwtClaimFromContext(ginContext)
		// Cek apakah user punya salah satu required permission
		if !hasPermission(userJwtClaim.Permissions, requiredPermissions) {
			ginContext.JSON(http.StatusUnauthorized, helper.NewErrorResponse("", helper.NewErrorDetail(
				exception.StatusAuthError, exception.ErrInsufficientPermission, nil)))
			ginContext.Abort()
			return
		}

		// lanjut ke next handler
		ginContext.Next()
	}
}

func hasPermission(userPermissions []string, requiredPermissions []string) bool {
	for _, required := range requiredPermissions {
		for _, userPerm := range userPermissions {
			if strings.EqualFold(userPerm, required) {
				return true
			}
		}
	}
	return false
}
