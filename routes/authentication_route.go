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
	authRouterGroup.GET("login", authenticationRoutes.userController.RenderLogin)
	authRouterGroup.GET("login-keycloak", authenticationRoutes.userController.Login)
	authRouterGroup.GET("callback", authenticationRoutes.userController.CallbackHandler)
}
