package routes

import (
	"zenith-on-stage/configs"
	"zenith-on-stage/internal/user"
	"zenith-on-stage/pkg/middleware"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

type ProtectedRoutes struct {
	viperConfig    *viper.Viper
	redisInstance  *configs.RedisInstance
	userController user.Controller
}

func NewProtectedRoutes(
	viperConfig *viper.Viper,
	redisInstance *configs.RedisInstance,
	userController user.Controller,

) *ProtectedRoutes {
	return &ProtectedRoutes{
		viperConfig: viperConfig,

		redisInstance:  redisInstance,
		userController: userController,
	}
}

func (protectedRoutes *ProtectedRoutes) Setup(routerGroup *gin.RouterGroup) {
	userRouterGroup := routerGroup.Group("users")
	userRouterGroup.GET("pagination", middleware.HasPermission("ROLE_USER_VIEW"), protectedRoutes.userController.FindAllUserPagination)
	userRouterGroup.GET("profile", protectedRoutes.userController.FindSelf)
	userRouterGroup.GET("/:id", middleware.HasPermission("ROLE_USER_VIEW"), protectedRoutes.userController.FindById)
	userRouterGroup.POST("", middleware.HasPermission("ROLE_USER_CREATE"), protectedRoutes.userController.CreateUser)
	userRouterGroup.PUT("/profile", protectedRoutes.userController.UpdateProfile)
	userRouterGroup.PUT("/:id", middleware.HasPermission("ROLE_USER_EDIT"), protectedRoutes.userController.UpdateUser)
	userRouterGroup.DELETE("/:id", middleware.HasPermission("ROLE_USER_DELETE"), protectedRoutes.userController.DeleteUser)
	userRouterGroup.GET("/logout", protectedRoutes.userController.LogoutUser)
}
