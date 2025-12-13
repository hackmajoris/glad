package dto

// Request DTOs

// RegisterRequest represents a user registration request
type RegisterRequest struct {
	Username string `json:"username" validate:"required,min=3,max=50"`
	Name     string `json:"name" validate:"required,min=2,max=100"`
	Password string `json:"password" validate:"required,min=6"`
}

// LoginRequest represents a user login request
type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// UpdateUserRequest represents a user update request
type UpdateUserRequest struct {
	Name     *string `json:"name,omitempty" validate:"omitempty,min=2,max=100"`
	Password *string `json:"password,omitempty" validate:"omitempty,min=6"`
}

// Response DTOs

// MessageResponse represents a simple message response
type MessageResponse struct {
	Message string `json:"message"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error"`
}

// TokenResponse represents a token response
type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
}

// ProtectedResponse represents a protected resource response
type ProtectedResponse struct {
	Message  string `json:"message"`
	Username string `json:"username"`
}

// UserListResponse represents a user in list responses (without password)
type UserListResponse struct {
	Username string `json:"username"`
	Name     string `json:"name"`
}

// CurrentUserResponse represents the current authenticated user's data
type CurrentUserResponse struct {
	Username  string `json:"username"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// Skill Request DTOs

// CreateSkillRequest represents a request to add a skill to a user
type CreateSkillRequest struct {
	SkillName         string `json:"skill_name" validate:"required,min=1,max=100"`
	ProficiencyLevel  string `json:"proficiency_level" validate:"required,oneof=Beginner Intermediate Advanced Expert"`
	YearsOfExperience int    `json:"years_of_experience" validate:"min=0"`
	Notes             string `json:"notes,omitempty" validate:"max=500"`
}

// UpdateSkillRequest represents a request to update a user's skill
type UpdateSkillRequest struct {
	ProficiencyLevel  *string `json:"proficiency_level,omitempty" validate:"omitempty,oneof=Beginner Intermediate Advanced Expert"`
	YearsOfExperience *int    `json:"years_of_experience,omitempty" validate:"omitempty,min=0"`
	Notes             *string `json:"notes,omitempty" validate:"omitempty,max=500"`
}

// Skill Response DTOs

// SkillResponse represents a skill in responses
type SkillResponse struct {
	SkillName         string `json:"skill_name"`
	ProficiencyLevel  string `json:"proficiency_level"`
	YearsOfExperience int    `json:"years_of_experience"`
	Endorsements      int    `json:"endorsements"`
	LastUsedDate      string `json:"last_used_date"`
	Notes             string `json:"notes,omitempty"`
	CreatedAt         string `json:"created_at"`
	UpdatedAt         string `json:"updated_at"`
}

// UserSkillResponse represents a user with a specific skill (for cross-user queries)
type UserSkillResponse struct {
	Username          string `json:"username"`
	Name              string `json:"name,omitempty"` // From GSI projection
	SkillName         string `json:"skill_name"`
	ProficiencyLevel  string `json:"proficiency_level"`
	YearsOfExperience int    `json:"years_of_experience"`
	Endorsements      int    `json:"endorsements"`
	LastUsedDate      string `json:"last_used_date"`
}

// Master Skill Request DTOs

// CreateMasterSkillRequest represents a request to create a master skill
type CreateMasterSkillRequest struct {
	SkillID     string   `json:"skill_id" validate:"required,min=1,max=50"`
	SkillName   string   `json:"skill_name" validate:"required,min=1,max=100"`
	Description string   `json:"description" validate:"max=500"`
	Category    string   `json:"category" validate:"required,min=1,max=50"`
	Tags        []string `json:"tags,omitempty"`
}

// UpdateMasterSkillRequest represents a request to update a master skill
type UpdateMasterSkillRequest struct {
	SkillName   string   `json:"skill_name,omitempty" validate:"omitempty,min=1,max=100"`
	Description string   `json:"description,omitempty" validate:"omitempty,max=500"`
	Category    string   `json:"category,omitempty" validate:"omitempty,min=1,max=50"`
	Tags        []string `json:"tags,omitempty"`
}

// Master Skill Response DTOs

// MasterSkillResponse represents a master skill in responses
type MasterSkillResponse struct {
	SkillID     string   `json:"skill_id"`
	SkillName   string   `json:"skill_name"`
	Description string   `json:"description"`
	Category    string   `json:"category"`
	Tags        []string `json:"tags,omitempty"`
	CreatedAt   string   `json:"created_at"`
	UpdatedAt   string   `json:"updated_at"`
}
