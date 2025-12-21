package models

import (
	"testing"
	"time"
)

func TestNewUser(t *testing.T) {
	tests := []struct {
		name     string
		username string
		userName string
		password string
		wantErr  bool
	}{
		{
			name:     "valid user creation",
			username: "testuser",
			userName: "Test User",
			password: "password123",
			wantErr:  false,
		},
		{
			name:     "empty username",
			username: "",
			userName: "Test User",
			password: "password123",
			wantErr:  true,
		},
		{
			name:     "empty name",
			username: "testuser",
			userName: "",
			password: "password123",
			wantErr:  true,
		},
		{
			name:     "empty password",
			username: "testuser",
			userName: "Test User",
			password: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := NewUser(tt.username, tt.userName, tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewUser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if user.Username != tt.username {
					t.Errorf("Expected username %s, got %s", tt.username, user.Username)
				}
				if user.Name != tt.userName {
					t.Errorf("Expected name %s, got %s", tt.userName, user.Name)
				}
				if user.PasswordHash == "" {
					t.Error("Expected password hash to be set")
				}
				if user.CreatedAt.IsZero() {
					t.Error("Expected CreatedAt to be set")
				}
				if user.UpdatedAt.IsZero() {
					t.Error("Expected UpdatedAt to be set")
				}
			}
		})
	}
}

func TestUser_ValidatePassword(t *testing.T) {
	user, err := NewUser("testuser", "Test User", "password123")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	tests := []struct {
		name     string
		password string
		want     bool
	}{
		{
			name:     "correct password",
			password: "password123",
			want:     true,
		},
		{
			name:     "incorrect password",
			password: "wrongpassword",
			want:     false,
		},
		{
			name:     "empty password",
			password: "",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := user.ValidatePassword(tt.password); got != tt.want {
				t.Errorf("User.ValidatePassword() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUser_UpdateName(t *testing.T) {
	user, err := NewUser("testuser", "Test User", "password123")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	originalUpdatedAt := user.UpdatedAt
	time.Sleep(10 * time.Millisecond) // Ensure time difference

	tests := []struct {
		name    string
		newName string
		wantErr bool
	}{
		{
			name:    "valid name update",
			newName: "Updated Name",
			wantErr: false,
		},
		{
			name:    "name too short",
			newName: "A",
			wantErr: true,
		},
		{
			name:    "name too long",
			newName: string(make([]byte, 101)),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userCopy := *user
			err := userCopy.UpdateName(tt.newName)
			if (err != nil) != tt.wantErr {
				t.Errorf("User.UpdateName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if userCopy.Name != tt.newName {
					t.Errorf("Expected name %s, got %s", tt.newName, userCopy.Name)
				}
				if userCopy.UpdatedAt.Equal(originalUpdatedAt) {
					t.Error("Expected UpdatedAt to be updated")
				}
			}
		})
	}
}

func TestUser_UpdatePassword(t *testing.T) {
	user, err := NewUser("testuser", "Test User", "password123")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	originalUpdatedAt := user.UpdatedAt
	time.Sleep(10 * time.Millisecond) // Ensure time difference

	tests := []struct {
		name        string
		newPassword string
		wantErr     bool
	}{
		{
			name:        "valid password update",
			newPassword: "newpassword123",
			wantErr:     false,
		},
		{
			name:        "password too short",
			newPassword: "123",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userCopy := *user
			err := userCopy.UpdatePassword(tt.newPassword)
			if (err != nil) != tt.wantErr {
				t.Errorf("User.UpdatePassword() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if !userCopy.ValidatePassword(tt.newPassword) {
					t.Error("Password was not updated correctly")
				}
				if userCopy.UpdatedAt.Equal(originalUpdatedAt) {
					t.Error("Expected UpdatedAt to be updated")
				}
			}
		})
	}
}

func TestUser_GetUsername(t *testing.T) {
	user, err := NewUser("testuser", "Test User", "password123")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	if got := user.GetUsername(); got != "testuser" {
		t.Errorf("User.GetUsername() = %v, want %v", got, "testuser")
	}
}
