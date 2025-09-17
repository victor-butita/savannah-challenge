package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gin-gonic/gin"
)

type OIDCVerifier interface {
	Verify(ctx context.Context, rawIDToken string) (*oidc.IDToken, error)
}

type AuthMiddleware struct {
	Verifier OIDCVerifier
}

func NewOIDCVerifier(providerURL, clientID string) (OIDCVerifier, error) {
	provider, err := oidc.NewProvider(context.Background(), providerURL)
	if err != nil {
		return nil, err
	}
	verifier := provider.Verifier(&oidc.Config{ClientID: clientID})
	return verifier, nil
}

func NewAuthMiddleware(verifier OIDCVerifier) *AuthMiddleware {
	return &AuthMiddleware{Verifier: verifier}
}

func (m *AuthMiddleware) ValidateToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid Authorization header format"})
			c.Abort()
			return
		}

		rawToken := parts[1]
		_, err := m.Verifier.Verify(c.Request.Context(), rawToken)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token", "details": err.Error()})
			c.Abort()
			return
		}

		c.Next()
	}
}
