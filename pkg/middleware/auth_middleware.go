package middleware

import (
	"net/http"
	"zenith-on-stage/configs"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gin-gonic/gin"
)

type AuthMiddleware struct {
	authClient   *configs.AuthenticationClient
	sessionStore configs.SessionStore
	clientID     string
}

// NewAuthMiddleware creates a new authentication middleware with OIDC verification
func NewAuthMiddleware(
	authClient *configs.AuthenticationClient,
	sessionStore configs.SessionStore,
) *AuthMiddleware {
	return &AuthMiddleware{
		authClient:   authClient,
		sessionStore: sessionStore,
	}
}
func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get session from cookie
		sessionID, err := c.Cookie("session_id")
		if err != nil {
			c.Redirect(http.StatusTemporaryRedirect, "/")
			c.Abort()
			return
		}
		// Get session data from Redis
		sessionData, err := m.sessionStore.Get(c, sessionID)
		if err != nil {
			// Clear invalid session cookie
			c.SetCookie("session_id", "", -1, "/", "", true, true)
			c.Redirect(http.StatusTemporaryRedirect, "/")
			c.Abort()
			return
		}
		// Verify the access token using the OIDC provider
		token, err := m.authClient.Provider.Verifier(&oidc.Config{
			SkipClientIDCheck: true, // Access tokens don't require client ID check
		}).Verify(c, sessionData.AccessToken)

		if err != nil {
			// The token is invalid - let's clean up and redirect
			m.sessionStore.Delete(c, sessionID)
			c.SetCookie("session_id", "", -1, "/", "", true, true)
			c.Redirect(http.StatusTemporaryRedirect, "/")
			c.Abort()
			return
		}
		// Extract claims from the token
		var claims map[string]interface{}
		if err := token.Claims(&claims); err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		// Store the validated claims and session in the context
		c.Set("user_session", sessionData)
		c.Set("user_claims", claims)
		c.Next()
	}
}
