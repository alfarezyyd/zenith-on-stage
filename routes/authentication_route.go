package routes

import (
	"zenith-on-stage/internal/user"

	"github.com/gin-gonic/gin"
)

type AuthenticationRoutes struct {
	userController user.Controller
}

func NewAuthenticationRoutes(userController user.Controller) *AuthenticationRoutes {
	return &AuthenticationRoutes{
		userController: userController,
	}
}

func (authenticationRoutes *AuthenticationRoutes) Setup(routerGroup *gin.RouterGroup) {
	authRouterGroup := routerGroup.Group("authentication")
	authRouterGroup.POST("login", authenticationRoutes.userController.Login)
}
