package user

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"time"
	"zenith-on-stage/configs"
	"zenith-on-stage/internal/entity"
	"zenith-on-stage/internal/model"
	"zenith-on-stage/internal/validator"
	"zenith-on-stage/pkg/exception"
	"zenith-on-stage/pkg/helper"
	"zenith-on-stage/pkg/logger"
	"zenith-on-stage/pkg/mapper"
	"zenith-on-stage/storage"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/spf13/viper"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
	"gorm.io/gorm"
)

type ServiceImpl struct {
	userRepository      Repository
	validatorService    validator.Service
	dbConnection        *gorm.DB
	viperConfig         *viper.Viper
	redisInstance       *configs.RedisInstance
	localStorageService *storage.Manager
	redisAuthManager    configs.AuthenticationStore
	authClient          *configs.AuthenticationClient
	sessionManager      configs.SessionStore
}

func NewService(userRepository Repository, validatorService validator.Service, dbConnection *gorm.DB,
	viperConfig *viper.Viper,
	redisInstance *configs.RedisInstance,
	localStorageService *storage.Manager,
	redisAuthManager configs.AuthenticationStore,
	authClient *configs.AuthenticationClient,
	sessionManager configs.SessionStore,
) *ServiceImpl {
	return &ServiceImpl{
		userRepository:      userRepository,
		validatorService:    validatorService,
		dbConnection:        dbConnection,
		viperConfig:         viperConfig,
		redisInstance:       redisInstance,
		localStorageService: localStorageService,
		redisAuthManager:    redisAuthManager,
		authClient:          authClient,
		sessionManager:      sessionManager,
	}
}

func (userService *ServiceImpl) FindAll() []*model.UserResponse {
	var allUser []*model.UserResponse
	err := userService.dbConnection.Transaction(func(gormTransaction *gorm.DB) error {
		userResponse, err := userService.userRepository.FindAll(gormTransaction)
		helper.CheckErrorOperation(err, exception.ParseGormError(err))
		allUser = helper.MapEntitiesIntoResponsesWithFunc[*entity.User, *model.UserResponse](userResponse, mapper.FuncMapAuditable)
		return nil
	})
	helper.CheckErrorOperation(err, exception.ParseGormError(err))
	return allUser
}

func (userService *ServiceImpl) FindAllPagination(paginationReq *model.PaginationRequest) *model.PaginatedResponse[*model.UserResponse] {
	paginationQuery := helper.BuildPaginationQuery(paginationReq)
	var userResponses []*model.UserResponse
	var totalItems int64

	err := userService.dbConnection.Transaction(func(gormTransaction *gorm.DB) error {
		userEntities, total, err := userService.userRepository.FindAllPagination(
			gormTransaction,
			paginationQuery.OrderClause,
			paginationQuery.Offset,
			paginationQuery.Limit,
			paginationQuery.SearchQuery,
		)
		helper.CheckErrorOperation(err, exception.ParseGormError(err))

		userResponses = helper.MapEntitiesIntoResponsesWithFunc[*entity.User, *model.UserResponse](
			userEntities,
			mapper.FuncMapAuditable,
		)
		totalItems = total

		helper.CheckErrorOperation(err, exception.ParseGormError(err))
		return nil
	})
	helper.CheckErrorOperation(err, exception.ParseGormError(err))
	return helper.NewPaginatedResponseFromResult(
		"Users fetched successfully",
		userResponses,
		paginationReq,
		totalItems,
	)
}

func (userService *ServiceImpl) FindById(ginContext *gin.Context, userId uint64) *model.UserResponse {
	var userResponse *model.UserResponse
	err := userService.dbConnection.Transaction(func(gormTransaction *gorm.DB) error {
		userEntity, err := userService.userRepository.FindById(gormTransaction, userId)
		helper.CheckErrorOperation(err, exception.ParseGormError(err))
		userResponse = helper.MapEntityIntoResponse[*entity.User, *model.UserResponse](userEntity,
			mapper.FuncMapAuditable)
		return nil
	})
	helper.CheckErrorOperation(err, exception.ParseGormError(err))
	return userResponse
}
func (userService *ServiceImpl) Login(ginContext *gin.Context) string {
	uniqueState, err := userService.generateRandomSecureString()
	helper.CheckErrorOperation(err, exception.NewApplicationError(http.StatusInternalServerError, exception.ErrInternalServerError))
	err = userService.redisAuthManager.SetState(ginContext.Request.Context(), uniqueState)
	helper.CheckErrorOperation(err, exception.NewApplicationError(http.StatusInternalServerError, exception.ErrInternalServerError))
	authURL := userService.authClient.OAuth.AuthCodeURL(
		uniqueState,
		oauth2.SetAuthURLParam("response_type", "code"),
		oauth2.SetAuthURLParam("scope", "openid profile email"),
	)
	return authURL
}

func (userService *ServiceImpl) FindSelf(ginContext *gin.Context) *model.UserResponse {
	userJwtClaims := helper.ExtractJwtClaimFromContext(ginContext)
	var userResponse *model.UserResponse
	err := userService.dbConnection.Transaction(func(gormTransaction *gorm.DB) error {
		userEntity, err := userService.userRepository.FindById(gormTransaction, userJwtClaims.Id)
		helper.CheckErrorOperation(err, exception.ParseGormError(err))
		userResponse = helper.MapEntityIntoResponse[*entity.User, *model.UserResponse](userEntity,
			mapper.FuncMapAuditable)
		return nil
	})
	helper.CheckErrorOperation(err, exception.ParseGormError(err))
	return userResponse
}

// Create - Membuat user baru
func (userService *ServiceImpl) Create(ginContext *gin.Context, createUserRequest *model.CreateUserRequest) *model.PaginatedResponse[*model.UserResponse] {
	userJwtClaims := helper.ExtractJwtClaimFromContext(ginContext)
	var paginationResp *model.PaginatedResponse[*model.UserResponse]
	valErr := userService.validatorService.ValidateStruct(createUserRequest)
	userService.validatorService.ParseValidationError(valErr, *createUserRequest)
	err := userService.dbConnection.Transaction(func(gormTransaction *gorm.DB) error {
		userEntity := helper.MapCreateRequestIntoEntity[model.CreateUserRequest, entity.User](createUserRequest)
		userEntity.Auditable = entity.NewAuditable(userJwtClaims.Name)
		err := userService.userRepository.Create(gormTransaction, userEntity)
		helper.CheckErrorOperation(err, exception.ParseGormError(err))

		return nil
	})
	helper.CheckErrorOperation(err, exception.ParseGormError(err))
	paginationRequest := model.NewPaginationRequest()
	paginationResp = userService.FindAllPagination(&paginationRequest)
	return paginationResp
}

func (userService *ServiceImpl) ValidateStateSession(ginContext *gin.Context, stateParam string) {
	// Retrieve stored state from Redis
	storedState, err := userService.redisAuthManager.GetState(ginContext, stateParam)
	helper.CheckErrorOperation(err, exception.NewApplicationError(http.StatusInternalServerError, fmt.Sprintf("failed to retrieve stored state: %s", stateParam)))
	if storedState != stateParam {
		helper.CheckErrorOperation(err, exception.NewApplicationError(http.StatusInternalServerError, fmt.Sprintf("state parameter mismatch")))
	}
	// Clean up used state from store
	if err = userService.redisAuthManager.DeleteState(ginContext, storedState); err != nil {
		logger.Error("Warning: failed to delete used state: %v", err)
	}
}

func (userService *ServiceImpl) TokenExchangeHandling(ginContext *gin.Context, authorizationCode string) *oauth2.Token {
	opts := []oauth2.AuthCodeOption{
		oauth2.SetAuthURLParam("grant_type", "authorization_code"),
	}
	oauth2Token, err := userService.authClient.OAuth.Exchange(ginContext, authorizationCode, opts...)
	helper.CheckErrorOperation(err, exception.NewApplicationError(http.StatusInternalServerError, "failed to exchange token"))
	return oauth2Token
}

func (userService *ServiceImpl) GetClaimsIdToken(ginContext *gin.Context, oAuth2Token *oauth2.Token) *model.OidcClaims {
	rawIDToken, isValid := oAuth2Token.Extra("id_token").(string)
	if !isValid {
		exception.ThrowApplicationError(exception.NewApplicationError(http.StatusInternalServerError, "Failed to get ID token"))
	}
	// Verify the ID token
	idToken, err := userService.authClient.OIDC.Verify(ginContext.Request.Context(), rawIDToken)
	if err != nil {
		exception.ThrowApplicationError(exception.NewApplicationError(http.StatusInternalServerError, "Failed to verify ID token"))
	}
	if idToken != nil {
		claims := model.OidcClaims{}
		if err := idToken.Claims(&claims); err != nil {
			exception.ThrowApplicationError(exception.NewApplicationError(http.StatusInternalServerError, "Failed to get user info"))
		}
		return &claims
	}

	exception.ThrowApplicationError(exception.NewApplicationError(http.StatusInternalServerError, "Failed to get user info"))
	return nil
}

func (userService *ServiceImpl) StoreSession(ginContext *gin.Context, oAuth2Token *oauth2.Token, oidcClaims *model.OidcClaims) {
	sessionID, err := userService.generateRandomSecureString()
	helper.CheckErrorOperation(err, exception.NewApplicationError(http.StatusInternalServerError, "error when generate session id"))
	sessionData := model.SessionData{
		AccessToken: oAuth2Token.AccessToken, // From Keycloak
		UserInfo: model.UserInfo{
			Username: oidcClaims.Email,
			Email:    oidcClaims.Email,
		},
		CreatedAt: time.Now(),
	}
	err = userService.sessionManager.Set(ginContext, sessionID, sessionData)
	helper.CheckErrorOperation(err, exception.NewApplicationError(http.StatusInternalServerError, "error when store session"))
	ginContext.SetSameSite(http.SameSiteStrictMode)
	ginContext.SetCookie(
		"session_id", // name
		sessionID,    // value
		3600,         // maxAge in seconds
		"/",          // path
		"",           // domain (empty means default to current domain)
		true,         // secure (HTTPS only)
		true,         // httpOnly (prevents JavaScript access)
	)

}

func (userService *ServiceImpl) HandleLogin(ginContext *gin.Context, loginUserRequest *model.LoginUserRequest) string {
	err := userService.validatorService.ValidateStruct(loginUserRequest)
	userService.validatorService.ParseValidationError(err, *loginUserRequest)
	var tokenString string
	err = userService.dbConnection.Transaction(func(gormTransaction *gorm.DB) error {
		userEntity, err := userService.userRepository.FindByIdentifier(gormTransaction, loginUserRequest.UserIdentifier)

		if err = bcrypt.CompareHashAndPassword([]byte(userEntity.Password), []byte(loginUserRequest.Password)); err != nil {
			exception.ThrowApplicationError(exception.NewApplicationError(http.StatusBadRequest, "User credentials invalid"))
		}
		jwtHour := userService.viperConfig.GetInt64("JWT_HOUR")
		if jwtHour == 0 {
			jwtHour = 72
		}
		tokenInstance := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"id":          userEntity.Id,
			"email":       userEntity.Email,
			"username":    userEntity.Username,
			"name":        userEntity.Name,
			"avatar_path": userEntity.AvatarPath,
			"exp":         time.Now().Add(time.Hour * time.Duration(jwtHour)).Unix(),
		})
		tokenString, err = tokenInstance.SignedString([]byte(userService.viperConfig.GetString("JWT_SECRET")))
		helper.CheckErrorOperation(err, exception.NewApplicationError(http.StatusInternalServerError, exception.ErrInternalServerError))
		if claims, ok := tokenInstance.Claims.(jwt.MapClaims); ok {
			userJwtClaim := helper.MapCreateRequestIntoEntity[jwt.MapClaims, model.JwtClaimRequest](&claims)
			ginContext.Set("claims", userJwtClaim)
		}
		redisKey := fmt.Sprintf("auth:token:%d", userEntity.Id)
		err = userService.redisInstance.RedisClient.Set(
			ginContext.Request.Context(),
			redisKey,
			tokenString,
			time.Hour*time.Duration(jwtHour),
		).Err()
		helper.CheckErrorOperation(err, exception.NewApplicationError(http.StatusInternalServerError, exception.ErrInternalServerError))

		return nil
	})
	return tokenString
}
func (userService *ServiceImpl) HandleLogout(ginContext *gin.Context) {
	userJwtClaims := helper.ExtractJwtClaimFromContext(ginContext)
	redisKey := fmt.Sprintf("auth:token:%d", userJwtClaims.Id)
	_, err := userService.redisInstance.RedisClient.Del(ginContext.Request.Context(), redisKey).Result()
	helper.CheckErrorOperation(err, exception.NewApplicationError(http.StatusInternalServerError, exception.ErrInternalServerError))
}
func (userService *ServiceImpl) Update(ginContext *gin.Context, updateUserRequest *model.UpdateUserRequest) *model.PaginatedResponse[*model.UserResponse] {
	var paginationResp *model.PaginatedResponse[*model.UserResponse]
	valErr := userService.validatorService.ValidateStruct(updateUserRequest)
	userService.validatorService.ParseValidationError(valErr, *updateUserRequest)
	err := userService.dbConnection.Transaction(func(gormTransaction *gorm.DB) error {
		userEntity, err := userService.userRepository.FindById(gormTransaction, updateUserRequest.Id)
		helper.CheckErrorOperation(err, exception.ParseGormError(err))
		helper.MapUpdateRequestIntoEntity(updateUserRequest, userEntity)
		if updateUserRequest.Password != "" {
			userEntity.Password, err = helper.HashPassword(updateUserRequest.Password)
			helper.CheckErrorOperation(err, exception.NewApplicationError(http.StatusInternalServerError, exception.ErrInternalServerError))
		}
		err = userService.userRepository.Update(gormTransaction, userEntity)
		helper.CheckErrorOperation(err, exception.ParseGormError(err))

		return nil
	})
	helper.CheckErrorOperation(err, exception.ParseGormError(err))
	paginationRequest := model.NewPaginationRequest()
	paginationResp = userService.FindAllPagination(&paginationRequest)
	return paginationResp
}

func (userService *ServiceImpl) UpdateProfile(ginContext *gin.Context, updateUserProfileRequest *model.UpdateUserProfileRequest) string {
	userJwtClaims := helper.ExtractJwtClaimFromContext(ginContext)
	updateUserProfileRequest.Id = userJwtClaims.Id
	valErr := userService.validatorService.ValidateStruct(updateUserProfileRequest)
	userService.validatorService.ParseValidationError(valErr, *updateUserProfileRequest)
	var tokenString string
	err := userService.dbConnection.Transaction(func(gormTransaction *gorm.DB) error {
		var err error
		userEntity, err := userService.userRepository.FindById(gormTransaction, updateUserProfileRequest.Id)
		helper.CheckErrorOperation(err, exception.ParseGormError(err))
		helper.MapUpdateRequestIntoEntity(updateUserProfileRequest, userEntity)
		if updateUserProfileRequest.Password != nil {
			userEntity.Password, _ = helper.HashPassword(*updateUserProfileRequest.Password)
		}
		if updateUserProfileRequest.Avatar != nil {
			newPath, err := userService.localStorageService.Disk().Put(updateUserProfileRequest.Avatar, fmt.Sprintf("users/profiles/%d-%s", time.Now().UnixNano(), updateUserProfileRequest.Avatar.Filename))
			helper.CheckErrorOperation(err, exception.NewApplicationError(http.StatusInternalServerError, exception.ErrSavingResources))
			userEntity.AvatarPath = newPath
		}

		err = userService.userRepository.Update(gormTransaction, userEntity)
		tokenString = userService.HandleLogin(ginContext, &model.LoginUserRequest{
			UserIdentifier: userEntity.Email,
			Password:       updateUserProfileRequest.CurrentPassword,
		})
		helper.CheckErrorOperation(err, exception.ParseGormError(err))
		return nil
	})
	helper.CheckErrorOperation(err, exception.ParseGormError(err))
	return tokenString
}

func (userService *ServiceImpl) Delete(ginContext *gin.Context, deleteUserRequest *model.DeleteResourceGeneralRequest) *model.PaginatedResponse[*model.UserResponse] {
	var paginationResp *model.PaginatedResponse[*model.UserResponse]
	valErr := userService.validatorService.ValidateStruct(deleteUserRequest)
	userService.validatorService.ParseValidationError(valErr, *deleteUserRequest)
	err := userService.dbConnection.Transaction(func(gormTransaction *gorm.DB) error {
		_, err := userService.userRepository.FindById(gormTransaction, deleteUserRequest.Id)
		helper.CheckErrorOperation(err, exception.ParseGormError(err))
		err = userService.userRepository.Delete(gormTransaction, deleteUserRequest.Id)
		helper.CheckErrorOperation(err, exception.ParseGormError(err))

		return nil
	})
	helper.CheckErrorOperation(err, exception.ParseGormError(err))
	paginationRequest := model.NewPaginationRequest()
	paginationResp = userService.FindAllPagination(&paginationRequest)
	return paginationResp
}

func (userService *ServiceImpl) generateRandomSecureString() (string, error) {
	arrOfByte := make([]byte, 32)
	if _, err := rand.Read(arrOfByte); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(arrOfByte), nil
}
