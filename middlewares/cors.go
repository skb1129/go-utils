package middlewares

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/skb1129/go-utils/config"
)

func CorsMiddleware() gin.HandlerFunc {
	origins := config.GetSlice("cors.origins")
	configHeaders := config.GetSlice("cors.headers")
	environment := config.GetString("environment")

	// Standard default headers that most APIs need
	defaultHeaders := []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"}
	allHeaders := append(defaultHeaders, configHeaders...)
	allowedHeaders := strings.Join(allHeaders, ", ")

	return func(c *gin.Context) {
		requestOrigin := c.Request.Header.Get("Origin")

		// If no Origin header, it's not a CORS request; skip middleware logic
		if requestOrigin == "" {
			c.Next()
			return
		}

		isAllowed := false
		for _, origin := range origins {
			if origin == "*" || requestOrigin == origin {
				isAllowed = true
				break
			}
		}

		if isAllowed || environment != "prod" {
			c.Header("Access-Control-Allow-Origin", requestOrigin)
		}

		c.Header("Access-Control-Allow-Headers", allowedHeaders)
		c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, DELETE, PUT, PATCH")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Max-Age", "600")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
