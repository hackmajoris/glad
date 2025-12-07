package main

import (
	"github.com/hackmajoris/glad/cmd/app/internal/database"
	"github.com/hackmajoris/glad/cmd/app/internal/handler"
	"github.com/hackmajoris/glad/cmd/app/internal/router"
	"github.com/hackmajoris/glad/cmd/app/internal/service"
	"github.com/hackmajoris/glad/pkg/auth"
	"github.com/hackmajoris/glad/pkg/config"
	"github.com/hackmajoris/glad/pkg/middleware"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize dependencies
	userRepo := database.NewDynamoDBRepository()
	tokenService := auth.NewTokenService(cfg)
	userService := service.NewUserService(userRepo, tokenService)
	apiHandler := handler.New(userService)
	authMiddleware := middleware.NewAuthMiddleware(tokenService)

	// Setup router
	r := setupRouter(apiHandler, authMiddleware)

	// Start Lambda
	lambda.Start(func(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		return r.Route(request)
	})
}

func setupRouter(h *handler.Handler, auth *middleware.AuthMiddleware) *router.Router {
	r := router.New()

	// Public routes
	r.POST("/register", h.Register)
	r.POST("/login", h.Login)

	// Protected routes
	r.GET("/protected", h.Protected, auth.RequireAuth())
	r.GET("/me", h.GetCurrentUser, auth.RequireAuth())
	r.PUT("/user", h.UpdateUser, auth.RequireAuth())
	r.GET("/users", h.ListUsers, auth.RequireAuth())

	return r
}
