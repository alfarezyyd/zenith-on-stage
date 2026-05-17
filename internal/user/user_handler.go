package user

import (
	"fmt"
	"net/http"
	"strconv"
	"zenith-on-stage/internal/model"
	"zenith-on-stage/pkg/exception"
	"zenith-on-stage/pkg/helper"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

type Handler struct {
	userService Service
	viperConfig *viper.Viper
}

func NewHandler(userService Service, viperConfig *viper.Viper) *Handler {
	return &Handler{
		userService: userService,
		viperConfig: viperConfig,
	}
}

func (userHandler *Handler) FindAllUser(ginContext *gin.Context) {
	userResponses := userHandler.userService.FindAll()
	ginContext.JSON(http.StatusOK, helper.WriteSuccess("User has been fetched", userResponses))
}

func (userHandler *Handler) Login(ginContext *gin.Context) {
	redirectURL := userHandler.userService.Login(ginContext)
	ginContext.Redirect(http.StatusTemporaryRedirect, redirectURL)
}

func (userHandler *Handler) FindById(ginContext *gin.Context) {
	userId := ginContext.Param("id")
	parsedUserId, err := strconv.ParseUint(userId, 10, 64)
	helper.CheckErrorOperation(err, exception.NewApplicationError(http.StatusBadRequest, exception.ErrParameterInvalid))
	userResponse := userHandler.userService.FindById(ginContext, parsedUserId)
	ginContext.JSON(http.StatusOK, helper.NewSuccessResponse("User fetched successfully", userResponse))
}

func (userHandler *Handler) CallbackHandler(ginContext *gin.Context) {
	stateParam := ginContext.Query("state")
	if stateParam == "" {
		exception.ThrowApplicationError(exception.NewApplicationError(http.StatusBadRequest, "Invalid or null state parameter"))
	}
	userHandler.userService.ValidateStateSession(ginContext, stateParam)
	authorizationCode := ginContext.Query("code")
	if authorizationCode == "" {
		exception.ThrowApplicationError(exception.NewApplicationError(http.StatusBadRequest, "authorizationCode is required"))
	}
	tokenExchangeHandling := userHandler.userService.TokenExchangeHandling(ginContext, authorizationCode)
	oidcClaims := userHandler.userService.GetClaimsIdToken(ginContext, tokenExchangeHandling)
	userHandler.userService.StoreSession(ginContext, tokenExchangeHandling, oidcClaims)
	ginContext.Redirect(http.StatusTemporaryRedirect, "/api/dashboard")
}

func (userHandler *Handler) RenderLogin(ginContext *gin.Context) {
	ginContext.HTML(http.StatusOK, "login.tmpl", gin.H{
		"Title":       "Welcome",
		"Message":     "Please log in to continue",
		"KeycloakURL": fmt.Sprintf("http://%s/api/authentication/login-keycloak", "localhost:8081"),
	})
}

func (userHandler *Handler) RenderDashboard(ginContext *gin.Context) {
	// Get session data with safe type assertion
	rawSession, exists := ginContext.Get("user_session")
	if !exists {
		ginContext.JSON(http.StatusInternalServerError, gin.H{"error": "No session found"})
		return
	}
	// Perform type assertion with error checking
	sessionData, ok := rawSession.(*model.SessionData)
	if !ok {
		ginContext.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid session data type"})
		return
	}
	// Now you can safely use the properly typed sessionData
	ginContext.HTML(http.StatusOK, "dashboard.tmpl", gin.H{
		"Title":        "Login Successful",
		"Message":      "You have successfully logged in!",
		"Username":     sessionData.UserInfo.Username,
		"Email":        sessionData.UserInfo.Email,
		"DashboardURL": "/dashboard",
		"LogoutURL":    "/logout",
	})
}

func (userHandler *Handler) FindSelf(ginContext *gin.Context) {
	userResponse := userHandler.userService.FindSelf(ginContext)
	ginContext.JSON(http.StatusOK, helper.NewSuccessResponse("User fetched successfully", userResponse))
}

func (userHandler *Handler) FindAllUserPagination(ginContext *gin.Context) {
	var paginationReq model.PaginationRequest
	err := ginContext.ShouldBindQuery(&paginationReq)
	helper.CheckErrorOperation(err, exception.NewApplicationError(http.StatusBadRequest, exception.ErrBadRequest))
	paginatedResponse := userHandler.userService.FindAllPagination(&paginationReq)
	ginContext.JSON(http.StatusOK, paginatedResponse)
}

func (userHandler *Handler) LoginUser(ginContext *gin.Context) {
	var loginUserRequest model.LoginUserRequest
	err := ginContext.ShouldBindBodyWithJSON(&loginUserRequest)
	helper.CheckErrorOperation(err, exception.NewApplicationError(http.StatusBadRequest, exception.ErrBadRequest))
	generatedToken := userHandler.userService.HandleLogin(ginContext, &loginUserRequest)
	ginContext.JSON(http.StatusOK, helper.NewSuccessResponse("User logged successfully", gin.H{
		"token": generatedToken,
	}))
}

func (userHandler *Handler) LogoutUser(ginContext *gin.Context) {
	userHandler.userService.HandleLogout(ginContext)
	ginContext.JSON(http.StatusOK, helper.NewSuccessResponse[interface{}]("User logout successfully", nil))
}

func (userHandler *Handler) CreateUser(ginContext *gin.Context) {
	var createUserModel model.CreateUserRequest
	err := ginContext.ShouldBindBodyWithJSON(&createUserModel)
	helper.CheckErrorOperation(err, exception.NewApplicationError(http.StatusBadRequest, exception.ErrBadRequest))
	paginatedResponse := userHandler.userService.Create(ginContext, &createUserModel)
	ginContext.JSON(http.StatusOK, paginatedResponse)
}

func (userHandler *Handler) UpdateUser(ginContext *gin.Context) {
	var updateUserModel model.UpdateUserRequest
	err := ginContext.ShouldBindBodyWithJSON(&updateUserModel)
	helper.CheckErrorOperation(err, exception.NewApplicationError(http.StatusBadRequest, exception.ErrBadRequest))
	userId := ginContext.Param("id")
	parsedUserId, err := strconv.ParseUint(userId, 10, 64)
	helper.CheckErrorOperation(err, exception.NewApplicationError(http.StatusBadRequest, exception.ErrBadRequest))
	updateUserModel.Id = parsedUserId
	paginatedResponse := userHandler.userService.Update(ginContext, &updateUserModel)
	ginContext.JSON(http.StatusOK, paginatedResponse)
}

func (userHandler *Handler) UpdateProfile(ginContext *gin.Context) {
	var updateUserProfileRequest model.UpdateUserProfileRequest
	err := ginContext.Request.ParseMultipartForm(20 << 20) // 32MB maxMemory
	helper.CheckErrorOperation(err, exception.NewApplicationError(http.StatusBadRequest, exception.ErrBadRequest))
	avatarFile, _ := ginContext.FormFile("avatar")
	updateUserProfileRequest.Avatar = avatarFile
	helper.CheckErrorOperation(err, exception.NewApplicationError(http.StatusBadRequest, exception.ErrBadRequest))
	newGeneratedToken := userHandler.userService.UpdateProfile(ginContext, &updateUserProfileRequest)
	ginContext.JSON(http.StatusOK, helper.NewSuccessResponse("Update profile success", gin.H{
		"token": newGeneratedToken,
	}))
}

func (userHandler *Handler) DeleteUser(ginContext *gin.Context) {
	var deleteUserModel model.DeleteResourceGeneralRequest
	err := ginContext.ShouldBindBodyWithJSON(&deleteUserModel)
	helper.CheckErrorOperation(err, exception.NewApplicationError(http.StatusBadRequest, exception.ErrBadRequest))
	userId := ginContext.Param("id")
	parsedUserId, err := strconv.ParseUint(userId, 10, 64)
	helper.CheckErrorOperation(err, exception.NewApplicationError(http.StatusBadRequest, exception.ErrBadRequest))
	deleteUserModel.Id = parsedUserId
	paginatedResponse := userHandler.userService.Delete(ginContext, &deleteUserModel)
	ginContext.JSON(http.StatusOK, paginatedResponse)
}
