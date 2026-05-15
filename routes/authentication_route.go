package routes

import (
	"github.com/gin-gonic/gin"
)

type AuthenticationRoutes struct {
}

func NewAuthenticationRoutes() *AuthenticationRoutes {
	return &AuthenticationRoutes{}
}

func (authenticationRoutes *AuthenticationRoutes) Setup(routerGroup *gin.RouterGroup) {
}
