package main

import (
	"log"

	"github.com/hackmajoris/glad-stack/cmd/glad/internal/database"
	"github.com/hackmajoris/glad-stack/cmd/glad/internal/handler"
	"github.com/hackmajoris/glad-stack/cmd/glad/internal/router"
	"github.com/hackmajoris/glad-stack/cmd/glad/internal/service"
	"github.com/hackmajoris/glad-stack/pkg/auth"
	"github.com/hackmajoris/glad-stack/pkg/config"
	"github.com/hackmajoris/glad-stack/pkg/middleware"

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
	corsMiddleware := middleware.NewCORSMiddleware([]string{"*"}) // Allow all origins for now

	// Setup router
	r := setupRouter(apiHandler, masterSkillHandler, authMiddleware, corsMiddleware)

	// Start Lambda
	lambda.Start(func(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		log.Println(request)
		return r.Route(request)
	})
}

func setupRouter(h *handler.Handler, msh *handler.MasterSkillHandler, auth *middleware.AuthMiddleware, cors *middleware.CORSMiddleware) *router.Router {
	r := router.New()

	// Apply CORS middleware globally
	r.Use(cors.AddCORSHeaders())

	// Protected routes - User Management
	// Note: Authentication is handled by API Gateway Cognito Authorizer
	r.GET("/me", h.GetCurrentUser)
	r.PUT("/user", h.UpdateUser)
	r.GET("/users", h.ListUsers)

	// Protected routes - Master Skill Management
	r.POST("/master-skills", msh.CreateMasterSkill)
	r.GET("/master-skills", msh.ListMasterSkills)
	r.GET("/master-skills/{skillID}", msh.GetMasterSkill)
	r.PUT("/master-skills/{skillID}", msh.UpdateMasterSkill)
	r.DELETE("/master-skills/{skillID}", msh.DeleteMasterSkill)

	// Protected routes - User Skill Management
	// Manage skills for a specific user
	r.POST("/users/{username}/skills", h.AddSkill)
	r.GET("/users/{username}/skills", h.ListSkillsForUser)
	r.GET("/users/{username}/skills/{skillName}", h.GetSkill)
	r.PUT("/users/{username}/skills/{skillName}", h.UpdateSkill)
	r.DELETE("/users/{username}/skills/{skillName}", h.DeleteSkill)

	// Query users by skill (cross-user queries using GSI)
	r.GET("/skills/{skillName}/users", h.ListUsersBySkill)

	return r
}
