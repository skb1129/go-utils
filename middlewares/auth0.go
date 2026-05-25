package middlewares

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/auth0/go-jwt-middleware/v3"
	"github.com/auth0/go-jwt-middleware/v3/jwks"
	"github.com/auth0/go-jwt-middleware/v3/validator"
	"github.com/gin-gonic/gin"
	"github.com/skb1129/go-utils/config"
	"github.com/skb1129/go-utils/logs"
	"github.com/skb1129/go-utils/request"
	"go.uber.org/zap"
)

// CustomClaims contains custom claims from the Auth0 token.
type CustomClaims struct {
	permission  string
	Permissions []string `json:"permissions"`
	Roles       []string `json:"app:roles"`
}

// Validate checks if the user has the required permission or is an ADMIN.
func (c *CustomClaims) Validate(ctx context.Context) error {
	// If the user has the ADMIN role, they have complete access.
	for _, role := range c.Roles {
		if role == "ADMIN" {
			return nil
		}
	}

	// If no specific permission is required, access is granted.
	if c.permission == "" {
		return nil
	}

	// Check if the required permission is in the user's permissions.
	for _, p := range c.Permissions {
		if p == c.permission {
			return nil
		}
	}

	return errors.New(string(request.PermissionDeniedError))
}

// Auth0Middleware validates Auth0 JWT tokens and puts the user ID and claims in the context.
func Auth0Middleware(permission string) gin.HandlerFunc {
	logger := logs.GetLogger()
	domain := config.GetString("auth0.domain")
	audience := config.GetString("auth0.audience")

	if domain == "" || audience == "" {
		logger.Fatal("Auth0 domain or audience not set in config")
	}

	issuerURL, err := url.Parse("https://" + domain + "/")
	if err != nil {
		logger.Fatal("Failed to parse Auth0 domain", zap.Error(err))
	}

	provider, err := jwks.NewCachingProvider(jwks.WithIssuerURL(issuerURL), jwks.WithCacheTTL(5*time.Minute))
	if err != nil {
		logger.Fatal("Failed to set up the JWKS provider", zap.Error(err))
	}

	jwtValidator, err := validator.New(
		validator.WithKeyFunc(provider.KeyFunc),
		validator.WithAlgorithm(validator.RS256),
		validator.WithIssuer(issuerURL.String()),
		validator.WithAudience(audience),
		validator.WithCustomClaims(func() validator.CustomClaims {
			return &CustomClaims{permission: permission}
		}),
		validator.WithAllowedClockSkew(time.Minute),
	)
	if err != nil {
		logger.Fatal("Failed to set up the Auth0 JWT validator", zap.Error(err))
	}

	return func(c *gin.Context) {
		extracted, err := jwtmiddleware.AuthHeaderTokenExtractor(c.Request)
		if err != nil || extracted.Token == "" {
			request.SendServiceError(c, request.CreateUnauthorizedError(fmt.Errorf("missing or invalid authorization header"), "Authorization token is required"))
			return
		}

		claims, err := jwtValidator.ValidateToken(c.Request.Context(), extracted.Token)
		if err != nil {
			if err.Error() == string(request.PermissionDeniedError) {
				request.SendServiceError(c, request.CreateForbiddenError(err, "Insufficient permissions"))
			} else {
				request.SendServiceError(c, request.CreateUnauthorizedError(err, "Invalid or expired token"))
			}
			return
		}

		validatedClaims := claims.(*validator.ValidatedClaims)
		customClaims := validatedClaims.CustomClaims.(*CustomClaims)
		isAdmin := false
		for _, role := range customClaims.Roles {
			if role == "ADMIN" {
				isAdmin = true
				break
			}
		}

		// Store user ID and claims in Gin context.
		c.Set("userID", validatedClaims.RegisteredClaims.Subject)
		c.Set("userClaims", customClaims)
		c.Set("isAdmin", isAdmin)

		// Also update request context so it's available in c.Request.Context().
		ctx := context.WithValue(c.Request.Context(), "userID", validatedClaims.RegisteredClaims.Subject)
		ctx = context.WithValue(ctx, "userClaims", customClaims)
		ctx = context.WithValue(ctx, "isAdmin", isAdmin)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}
