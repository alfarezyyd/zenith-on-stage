package routes

import (
	"zenith-on-stage/configs"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

type ProtectedRoutes struct {
	viperConfig   *viper.Viper
	redisInstance *configs.RedisInstance
}

func NewProtectedRoutes(
	viperConfig *viper.Viper,
	redisInstance *configs.RedisInstance,

) *ProtectedRoutes {
	return &ProtectedRoutes{
		viperConfig: viperConfig,

		redisInstance: redisInstance,
	}
}

func (protectedRoutes *ProtectedRoutes) Setup(routerGroup *gin.RouterGroup) {
}
