package user

import (
	"zenith-on-stage/internal/model"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
)

type Service interface {
	FindAll() []*model.UserResponse
	FindById(ginContext *gin.Context, userId uint64) *model.UserResponse
	Login(ginContext *gin.Context) string
	ValidateStateSession(ginContext *gin.Context, stateParam string)
	FindAllPagination(paginationReq *model.PaginationRequest) *model.PaginatedResponse[*model.UserResponse]
	HandleLogin(ginContext *gin.Context, loginUserRequest *model.LoginUserRequest) string
	HandleLogout(ginContext *gin.Context)
	Create(ginContext *gin.Context, createUserRequest *model.CreateUserRequest) *model.PaginatedResponse[*model.UserResponse]
	Update(ginContext *gin.Context, updateUserRequest *model.UpdateUserRequest) *model.PaginatedResponse[*model.UserResponse]
	UpdateProfile(ginContext *gin.Context, updateUserProfileRequest *model.UpdateUserProfileRequest) string
	Delete(ginContext *gin.Context, deleteUserRequest *model.DeleteResourceGeneralRequest) *model.PaginatedResponse[*model.UserResponse]
	FindSelf(ginContext *gin.Context) *model.UserResponse
	TokenExchangeHandling(ginContext *gin.Context, authorizationCode string) *oauth2.Token
	GetClaimsIdToken(ginContext *gin.Context, oAuth2Token *oauth2.Token) *model.OidcClaims
	StoreSession(ginContext *gin.Context, oAuth2Token *oauth2.Token, oidcClaims *model.OidcClaims)
}
