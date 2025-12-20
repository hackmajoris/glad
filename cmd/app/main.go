package main

import (
	"log"

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
	repo := database.NewRepository(cfg)
	tokenService := auth.NewTokenService(cfg)

	// Initialize services
	userService := service.NewUserService(repo, tokenService)
	skillService := service.NewSkillService(repo, repo, repo) // repo implements SkillRepository, MasterSkillRepository, and UserRepository
	masterSkillService := service.NewMasterSkillService(repo)

	// Initialize handlers
	apiHandler := handler.New(userService, skillService)
	masterSkillHandler := handler.NewMasterSkillHandler(masterSkillService)
	authMiddleware := middleware.NewAuthMiddleware(tokenService)

	// Setup router
	r := setupRouter(apiHandler, masterSkillHandler, authMiddleware)

	// Start Lambda
	lambda.Start(func(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		log.Println(request)
		return r.Route(request)
	})
}

func setupRouter(h *handler.Handler, msh *handler.MasterSkillHandler, auth *middleware.AuthMiddleware) *router.Router {
	r := router.New()

	// Public routes
	r.POST("/register", h.Register)
	r.POST("/login", h.Login)

	// Protected routes - User Management
	r.GET("/protected", h.Protected, auth.RequireAuth())
	r.GET("/me", h.GetCurrentUser, auth.RequireAuth())
	r.PUT("/user", h.UpdateUser, auth.RequireAuth())
	r.GET("/users", h.ListUsers, auth.RequireAuth())

	// Protected routes - Master Skill Management
	r.POST("/master-skills", msh.CreateMasterSkill, auth.RequireAuth())
	r.GET("/master-skills", msh.ListMasterSkills, auth.RequireAuth())
	r.GET("/master-skills/{skillID}", msh.GetMasterSkill, auth.RequireAuth())
	r.PUT("/master-skills/{skillID}", msh.UpdateMasterSkill, auth.RequireAuth())
	r.DELETE("/master-skills/{skillID}", msh.DeleteMasterSkill, auth.RequireAuth())

	// Protected routes - User Skill Management
	// Manage skills for a specific user
	r.POST("/users/{username}/skills", h.AddSkill, auth.RequireAuth())
	r.GET("/users/{username}/skills", h.ListSkillsForUser, auth.RequireAuth())
	r.GET("/users/{username}/skills/{skillName}", h.GetSkill, auth.RequireAuth())
	r.PUT("/users/{username}/skills/{skillName}", h.UpdateSkill, auth.RequireAuth())
	r.DELETE("/users/{username}/skills/{skillName}", h.DeleteSkill, auth.RequireAuth())

	// Query users by skill (cross-user queries using GSI)
	r.GET("/skills/{skillName}/users", h.ListUsersBySkill, auth.RequireAuth())

	return r
}
