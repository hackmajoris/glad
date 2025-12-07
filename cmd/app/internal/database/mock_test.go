package database

import (
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
	}
	if repo.users == nil {
		t.Error("Expected users map to be initialized")
	}
	if len(repo.users) != 0 {
		t.Error("Expected empty users map")
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
	if err != apperrors.ErrUserExists {
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
	repo.CreateUser(user)

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
	if err != apperrors.ErrUserNotFound {
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
	repo.CreateUser(user)

	// Update user
	user.UpdateName("Updated User")

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
	if err != apperrors.ErrUserNotFound {
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
	repo.CreateUser(user)

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

	repo.CreateUser(user1)
	repo.CreateUser(user2)

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
			repo.CreateUser(user)
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
