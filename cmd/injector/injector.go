package injector

import (
	"fmt"
	"time"
	"zenith-on-stage/configs"
	"zenith-on-stage/internal/user"
	validatorService "zenith-on-stage/internal/validator"
	"zenith-on-stage/pkg/exception"
	"zenith-on-stage/pkg/logger"
	"zenith-on-stage/pkg/middleware"
	"zenith-on-stage/routes"
	"zenith-on-stage/storage"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	universalTranslator "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	"go.uber.org/fx"
	"gorm.io/gorm"
)

func NewDatabaseConnection(databaseCredentials *configs.DatabaseCredentials) *gorm.DB {
	databaseConnectionInstance := configs.NewDatabaseConnection(databaseCredentials)
	return databaseConnectionInstance.GetDatabaseConnection()
}

func NewRedisInstance(redisConfig configs.RedisConfig) *configs.RedisInstance {
	redisInstance, err := configs.InitRedisInstance(redisConfig)
	if err != nil {
		panic(err)
	}
	return redisInstance
}

func NewRedisClient(redisInstance *configs.RedisInstance) *redis.Client {
	return redisInstance.RedisClient
}

func InitRedisConfig(viperConfig *viper.Viper) configs.RedisConfig {
	redisHostName, redisPassword, parsedRedisDatabaseIndex := configs.LoadRedisConfigFromEnvironment(viperConfig)
	return configs.NewRedisConfig(redisHostName, redisPassword, int(parsedRedisDatabaseIndex))
}

func NewValidator(gormDatabase *gorm.DB) (*validator.Validate, universalTranslator.Translator) {
	return configs.InitializeValidator(gormDatabase)
}

// NewViperConfig --- Provider untuk Viper config ---
func NewViperConfig() *viper.Viper {
	viperConfig := viper.New()
	viperConfig.AutomaticEnv()
	viperConfig.SetConfigType("env")
	// dev, prod, staging
	hostEnvironment := viperConfig.Get("HOST_ENVIRONMENT")
	viperConfig.SetConfigFile(fmt.Sprintf(".%s-env", hostEnvironment))

	if err := viperConfig.ReadInConfig(); err != nil {
		panic(err)
	}
	return viperConfig
}

// NewDatabaseCredentials --- Provider untuk Database Credentials ---
func NewDatabaseCredentials(viperConfig *viper.Viper) *configs.DatabaseCredentials {
	return &configs.DatabaseCredentials{
		DatabaseHost:     viperConfig.GetString("DATABASE_HOST"),
		DatabasePort:     viperConfig.GetString("DATABASE_PORT"),
		DatabaseName:     viperConfig.GetString("DATABASE_NAME"),
		DatabasePassword: viperConfig.GetString("DATABASE_PASSWORD"),
		DatabaseUsername: viperConfig.GetString("DATABASE_USERNAME"),
	}
}

func NewAuthenticationConfig(viperConfig *viper.Viper) *configs.AuthenticationConfig {
	return &configs.AuthenticationConfig{
		BaseURL:      viperConfig.GetString("KEYCLOAK_URL"),
		ClientID:     viperConfig.GetString("KEYCLOAK_CLIENT_ID"),
		RedirectURL:  viperConfig.GetString("KEYCLOAK_REDIRECT_URL"),
		ClientSecret: viperConfig.GetString("KEYCLOAK_CLIENT_SECRET"),
		Realm:        viperConfig.GetString("KEYCLOAK_REALM"),
	}
}

// NewGinEngine --- Provider untuk Gin Engine ---
func NewGinEngine() (*gin.Engine, *gin.RouterGroup) {
	gin.SetMode(gin.DebugMode)
	ginEngine := gin.Default()
	ginEngine.LoadHTMLGlob("internal/templates/*.*")
	ginEngine.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"*"},
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{"*"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	ginEngine.Use(gin.Recovery())
	ginEngine.Use(exception.Interceptor())
	ginEngineRoot := ginEngine.Group("/")
	ginEngineRoot.Use(middleware.RequestMetaMiddleware())

	return ginEngine, ginEngineRoot
}

var CoreModule = fx.Module("coreModule", fx.Provide(
	NewViperConfig,
	NewDatabaseCredentials,
	NewDatabaseConnection,
	NewValidator,
	NewGinEngine,
	InitRedisConfig,
	NewRedisInstance,
	NewStorageManager,
	NewAuthenticationConfig,
	NewAuthenticationClient,
	NewRedisAuthManager,
	NewSessionManager,
	NewRedisClient,
	NewAuthMiddleware,
))

var ApplicationRoutesModule = fx.Module("applicationRoutes",
	fx.Provide(
		routes.NewPublicRoutes,
		routes.NewAuthenticationRoutes,
		routes.NewProtectedRoutes,
		func(
			ginEngine *gin.Engine,
			publicRoutes *routes.PublicRoutes,
			authenticationRoutes *routes.AuthenticationRoutes,
			protectedRoutes *routes.ProtectedRoutes,
		) *routes.ApplicationRoutes {
			return routes.NewApplicationRoutes(ginEngine, publicRoutes, authenticationRoutes, protectedRoutes)
		},
	),
	fx.Invoke(func(applicationRoutes *routes.ApplicationRoutes) {
		applicationRoutes.Setup()
	}),
)

func NewStorageManager(viperConfig *viper.Viper) *storage.Manager {
	storageConfig := configs.NewStorageConfig(viperConfig)
	storageManager, err := storage.NewStorageManager(storageConfig)
	if err != nil {
		panic(err)
	}
	return storageManager
}

func NewAuthenticationClient(authConfig *configs.AuthenticationConfig) *configs.AuthenticationClient {
	authenticationClient, err := configs.NewAuthenticationClient(authConfig)
	if err != nil {
		logger.WithError(err)
	}
	return authenticationClient
}

func NewRedisAuthManager(redisClient *redis.Client) configs.AuthenticationStore {
	return configs.NewRedisAuthManager(redisClient)
}

func NewSessionManager(redisClient *redis.Client) configs.SessionStore {
	return configs.NewSessionRedisManager(redisClient)
}

func NewAuthMiddleware(authClient *configs.AuthenticationClient,
	sessionStore configs.SessionStore) *middleware.AuthMiddleware {
	return middleware.NewAuthMiddleware(authClient, sessionStore)
}

var UserModule = fx.Module("userFeature",
	fx.Provide(fx.Annotate(user.NewRepository, fx.As(new(user.Repository)))),
	fx.Provide(fx.Annotate(user.NewService, fx.As(new(user.Service)))),
	fx.Provide(fx.Annotate(user.NewHandler, fx.As(new(user.Controller)))),
)

var ValidatorModule = fx.Module("validatorFeature",
	fx.Provide(fx.Annotate(validatorService.NewService, fx.As(new(validatorService.Service)))),
)
