# Go Utils

A collection of Go utilities for building robust and scalable applications with Gin, Redis, MongoDB, Zap, and Viper.

## Features

- **Cache:** Redis client wrapper with support for JSON, Hashing, Lists, Sets, and Distributed Locking.
- **Config:** Flexible configuration management using Viper, supporting environment variables and JSON files.
- **Database:** Easy MongoDB client initialization.
- **Logging:** Structured logging using Uber's Zap.
- **Middlewares:**
  - `GinLogger`: Structured request logging.
  - `GinRecovery`: Panic recovery with error reporting.
  - `CorsMiddleware`: Configurable Cross-Origin Resource Sharing.
  - `Auth0Middleware`: JWT validation and scope checking.
  - `RateLimiter`: Redis-based rate limiting.
  - `ParseUserAgent`: User-agent parsing.
- **Request/Response:** Standardized Gin response helpers and request validation using `validator/v10`.

## Installation

```bash
go get github.com/skb1129/go-utils
```

## Core Modules

### Configuration

Initialize the configuration from a JSON string in an environment variable (`CONFIG`) or a local JSON file.

```go
import "github.com/skb1129/go-utils/config"

// Initialize
config.Init()

// Usage
dbUser := config.GetString("mongodb.user")
port := config.GetInt("server.port")
```

### Logging

Structured logging with different profiles for development and production.

```go
import "github.com/skb1129/go-utils/logs"

logger := logs.GetLogger()
logger.Info("Starting application", zap.String("version", "1.0.0"))
```

### Cache (Redis)

A wrapper around `go-redis` with convenient methods.

```go
import "github.com/skb1129/go-utils/cache"

r := cache.NewCache()
ctx := context.Background()

// Set and Get JSON
err := r.SetJSON(ctx, "user:1", user, time.Hour)
err = r.GetJSON(ctx, "user:1", &user)

// Locking
locked, err := r.Lock(ctx, "resource_lock", time.Minute)
```

### Database (MongoDB)

Simple MongoDB client setup.

```go
import "github.com/skb1129/go-utils/db"

client := db.InitMongoDB()
collection := client.Database("mydb").Collection("mycoll")
```

### Middlewares

Enhance your Gin router with pre-built middlewares.

```go
import (
    "github.com/gin-gonic/gin"
    "github.com/skb1129/go-utils/middlewares"
)

r := gin.New()
r.Use(middlewares.GinLogger())
r.Use(middlewares.GinRecovery())

// Auth0 Protection
r.GET("/secure", middlewares.Auth0Middleware("read:data"), handler)

// CORS
r.Use(middlewares.CorsMiddleware())

// Rate Limiting
r.Use(middlewares.RateLimiter(cacheClient, "rate_limit:%s"))
```

### Request & Response Helpers

Standardize your API responses.

```go
import (
    "github.com/skb1129/go-utils/request"
)

func Handler(c *gin.Context) {
    var req MyRequest
    if err := request.ValidateRequest(c, &req); err != nil {
        request.SendServiceError(c, err)
        return
    }

    // ... logic ...
    
    request.SendSuccessResponse(c, result)
}
```

## Environment Variables

- `ENV_FILE`: If set to `prod`, the library looks for `./env.prod.json`.
- `CONFIG`: If `ENV_FILE` is not set, it reads the full JSON configuration from this variable.

### CORS Configuration

The `CorsMiddleware` expects the following configuration:

- `cors.origins`: A slice of allowed origins.
- `cors.headers`: A slice of allowed headers.
- `environment`: If not set to `prod`, `Access-Control-Allow-Origin` will reflect the request's origin even if not in the allowed list.

## License

MIT
