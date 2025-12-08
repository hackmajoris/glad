package database

import (
	"errors"
	"fmt"
	"sync"
	"testing"

	apperrors "github.com/hackmajoris/glad/cmd/app/internal/errors"
	"github.com/hackmajoris/glad/cmd/app/internal/models"
)

func TestNewMockRepository(t *testing.T) {
	repo := NewMockRepository()
	if repo == nil {
		t.Error("Expected non-nil repository")
		return
	}

	if repo.users == nil {
		t.Error("Expected users map to be initialized")
	}
	if repo.skills == nil {
		t.Error("Expected skills map to be initialized")
	}
	if len(repo.users) != 0 {
		t.Error("Expected empty users map")
	}
	if len(repo.skills) != 0 {
		t.Error("Expected empty skills map")
	}
}

func TestMockRepository_CreateUser(t *testing.T) {
	repo := NewMockRepository()

	user, err := models.NewUser("testuser", "Test User", "password123")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Test successful creation
	err = repo.CreateUser(user)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Test duplicate creation
	err = repo.CreateUser(user)
	if !errors.Is(err, apperrors.ErrUserExists) {
		t.Errorf("Expected ErrUserExists, got %v", err)
	}
}

func TestMockRepository_GetUser(t *testing.T) {
	repo := NewMockRepository()

	user, err := models.NewUser("testuser", "Test User", "password123")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Create user first
	err = repo.CreateUser(user)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Test successful retrieval
	retrieved, err := repo.GetUser("testuser")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if retrieved.Username != "testuser" {
		t.Errorf("Expected username testuser, got %s", retrieved.Username)
	}
	if retrieved.Name != "Test User" {
		t.Errorf("Expected name 'Test User', got %s", retrieved.Name)
	}

	// Test non-existent user
	_, err = repo.GetUser("nonexistent")
	if !errors.Is(err, apperrors.ErrUserNotFound) {
		t.Errorf("Expected ErrUserNotFound, got %v", err)
	}
}

func TestMockRepository_UpdateUser(t *testing.T) {
	repo := NewMockRepository()

	user, err := models.NewUser("testuser", "Test User", "password123")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Create user first
	err = repo.CreateUser(user)
	if err != nil {
		return
	}

	// Update user
	err = user.UpdateName("Updated User")
	if err != nil {
		return
	}

	// Test successful update
	err = repo.UpdateUser(user)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Verify update
	retrieved, _ := repo.GetUser("testuser")
	if retrieved.Name != "Updated User" {
		t.Errorf("Expected name 'Updated User', got %s", retrieved.Name)
	}

	// Test update non-existent user
	nonExistentUser, _ := models.NewUser("nonexistent", "Non Existent", "password123")

	err = repo.UpdateUser(nonExistentUser)
	if !errors.Is(err, apperrors.ErrUserNotFound) {
		t.Errorf("Expected ErrUserNotFound, got %v", err)
	}
}

func TestMockRepository_UserExists(t *testing.T) {
	repo := NewMockRepository()

	user, err := models.NewUser("testuser", "Test User", "password123")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Test non-existent user
	exists, err := repo.UserExists("testuser")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if exists {
		t.Error("Expected user to not exist")
	}

	// Create user
	err = repo.CreateUser(user)
	if err != nil {
		return
	}

	// Test existing user
	exists, err = repo.UserExists("testuser")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if !exists {
		t.Error("Expected user to exist")
	}
}

func TestMockRepository_ListUsers(t *testing.T) {
	repo := NewMockRepository()

	// Test empty list
	users, err := repo.ListUsers()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if len(users) != 0 {
		t.Errorf("Expected empty list, got %d users", len(users))
	}

	// Create multiple users
	user1, _ := models.NewUser("user1", "User One", "password123")
	user2, _ := models.NewUser("user2", "User Two", "password123")

	err = repo.CreateUser(user1)
	if err != nil {
		return
	}

	err = repo.CreateUser(user2)
	if err != nil {
		return
	}

	// Test list with users
	users, err = repo.ListUsers()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if len(users) != 2 {
		t.Errorf("Expected 2 users, got %d", len(users))
	}

	// Verify users are in the list
	usernames := make(map[string]bool)
	for _, user := range users {
		usernames[user.Username] = true
	}
	if !usernames["user1"] {
		t.Error("Expected user1 to be in the list")
	}
	if !usernames["user2"] {
		t.Error("Expected user2 to be in the list")
	}
}

func TestMockRepository_ConcurrentAccess(t *testing.T) {
	repo := NewMockRepository()
	var wg sync.WaitGroup
	concurrency := 10

	// Test concurrent writes
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			user, _ := models.NewUser(
				fmt.Sprintf("user%d", id),
				fmt.Sprintf("User %d", id),
				"password123",
			)

			err := repo.CreateUser(user)
			if err != nil {
				return
			}
		}(i)
	}

	wg.Wait()

	// Verify all users were created
	users, _ := repo.ListUsers()
	if len(users) != concurrency {
		t.Errorf("Expected %d users, got %d", concurrency, len(users))
	}

	// Test concurrent reads
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			username := fmt.Sprintf("user%d", id)
			_, err := repo.GetUser(username)
			if err != nil {
				t.Errorf("Failed to get user %s: %v", username, err)
			}
		}(i)
	}

	wg.Wait()
}

// ============================================================================
// SKILL REPOSITORY TESTS
// ============================================================================

func TestMockRepository_CreateSkill(t *testing.T) {
	repo := NewMockRepository()

	skill, err := models.NewUserSkill("testuser", "Go", models.ProficiencyIntermediate, 3)
	if err != nil {
		t.Fatalf("Failed to create skill: %v", err)
	}

	// Test successful creation
	err = repo.CreateSkill(skill)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Test duplicate creation
	err = repo.CreateSkill(skill)
	if err != apperrors.ErrSkillAlreadyExists {
		t.Errorf("Expected ErrSkillAlreadyExists, got %v", err)
	}
}

func TestMockRepository_GetSkill(t *testing.T) {
	repo := NewMockRepository()

	skill, err := models.NewUserSkill("testuser", "Go", models.ProficiencyIntermediate, 3)
	if err != nil {
		t.Fatalf("Failed to create skill: %v", err)
	}

	// Create skill first
	repo.CreateSkill(skill)

	// Test successful retrieval
	retrieved, err := repo.GetSkill("testuser", "Go")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if retrieved.Username != "testuser" {
		t.Errorf("Expected username testuser, got %s", retrieved.Username)
	}
	if retrieved.SkillName != "Go" {
		t.Errorf("Expected skill name Go, got %s", retrieved.SkillName)
	}

	// Test non-existent skill
	_, err = repo.GetSkill("testuser", "nonexistent")
	if err != apperrors.ErrSkillNotFound {
		t.Errorf("Expected ErrSkillNotFound, got %v", err)
	}
}

func TestMockRepository_ListSkillsForUser(t *testing.T) {
	repo := NewMockRepository()

	// Test empty list
	skills, err := repo.ListSkillsForUser("testuser")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if len(skills) != 0 {
		t.Errorf("Expected empty list, got %d skills", len(skills))
	}

	// Create multiple skills for the same user
	skill1, _ := models.NewUserSkill("testuser", "Go", models.ProficiencyIntermediate, 3)
	skill2, _ := models.NewUserSkill("testuser", "Python", models.ProficiencyAdvanced, 5)
	skill3, _ := models.NewUserSkill("otheruser", "Java", models.ProficiencyBeginner, 1)

	repo.CreateSkill(skill1)
	repo.CreateSkill(skill2)
	repo.CreateSkill(skill3)

	// Test list skills for testuser
	skills, err = repo.ListSkillsForUser("testuser")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if len(skills) != 2 {
		t.Errorf("Expected 2 skills, got %d", len(skills))
	}

	// Verify correct skills are returned
	skillNames := make(map[string]bool)
	for _, skill := range skills {
		if skill.Username != "testuser" {
			t.Errorf("Expected username testuser, got %s", skill.Username)
		}
		skillNames[skill.SkillName] = true
	}
	if !skillNames["Go"] {
		t.Error("Expected Go skill to be in the list")
	}
	if !skillNames["Python"] {
		t.Error("Expected Python skill to be in the list")
	}
	if skillNames["Java"] {
		t.Error("Did not expect Java skill to be in the list for testuser")
	}
}

func TestMockRepository_ListUsersBySkill(t *testing.T) {
	repo := NewMockRepository()

	// Create skills for different users with same skill name
	skill1, _ := models.NewUserSkill("user1", "Go", models.ProficiencyIntermediate, 3)
	skill2, _ := models.NewUserSkill("user2", "Go", models.ProficiencyAdvanced, 5)
	skill3, _ := models.NewUserSkill("user3", "Python", models.ProficiencyBeginner, 1)

	repo.CreateSkill(skill1)
	repo.CreateSkill(skill2)
	repo.CreateSkill(skill3)

	// Test list users with Go skill
	skills, err := repo.ListUsersBySkill("Go")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if len(skills) != 2 {
		t.Errorf("Expected 2 users with Go skill, got %d", len(skills))
	}

	// Verify correct users are returned
	usernames := make(map[string]bool)
	for _, skill := range skills {
		if skill.SkillName != "Go" {
			t.Errorf("Expected skill name Go, got %s", skill.SkillName)
		}
		usernames[skill.Username] = true
	}
	if !usernames["user1"] {
		t.Error("Expected user1 to be in the list")
	}
	if !usernames["user2"] {
		t.Error("Expected user2 to be in the list")
	}
	if usernames["user3"] {
		t.Error("Did not expect user3 to be in the list for Go skill")
	}
}

func TestMockRepository_UnifiedInterface(t *testing.T) {
	// Test that the same repository instance implements both interfaces
	repo := NewMockRepository()

	// Test as UserRepository
	var userRepo UserRepository = repo
	user, _ := models.NewUser("testuser", "Test User", "password123")
	err := userRepo.CreateUser(user)
	if err != nil {
		t.Errorf("Failed to create user via UserRepository interface: %v", err)
	}

	// Test as SkillRepository
	var skillRepo SkillRepository = repo
	skill, _ := models.NewUserSkill("testuser", "Go", models.ProficiencyIntermediate, 3)
	err = skillRepo.CreateSkill(skill)
	if err != nil {
		t.Errorf("Failed to create skill via SkillRepository interface: %v", err)
	}

	// Test as combined Repository interface
	var combinedRepo Repository = repo
	_, err = combinedRepo.GetUser("testuser")
	if err != nil {
		t.Errorf("Failed to get user via Repository interface: %v", err)
	}
	_, err = combinedRepo.GetSkill("testuser", "Go")
	if err != nil {
		t.Errorf("Failed to get skill via Repository interface: %v", err)
	}
}
