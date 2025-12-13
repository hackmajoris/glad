package database

import (
	"fmt"
	"strings"
)

// Entity ID utility functions for consistent key generation across the application.
// All entity IDs use "#" as the delimiter for better DynamoDB practices.

// BuildUserEntityID creates an entity ID for a User
// Format: USER#<username>
func BuildUserEntityID(username string) string {
	return fmt.Sprintf("USER#%s", strings.ToLower(username))
}

// BuildUserSkillEntityID creates an entity ID for a UserSkill
// Format: USERSKILL#<username>#<skillID>
func BuildUserSkillEntityID(username, skillID string) string {
	return fmt.Sprintf("USERSKILL#%s#%s", strings.ToLower(username), strings.ToLower(skillID))
}

// BuildMasterSkillEntityID creates an entity ID for a MasterSkill
// Format: SKILL#<skillID>
func BuildMasterSkillEntityID(skillID string) string {
	return fmt.Sprintf("SKILL#%s", strings.ToLower(skillID))
}

// ParseUserEntityID extracts the username from a User entity ID
// Returns the username or empty string if invalid format
func ParseUserEntityID(entityID string) string {
	parts := strings.Split(entityID, "#")
	if len(parts) == 2 && parts[0] == "USER" {
		return parts[1]
	}
	return ""
}

// ParseUserSkillEntityID extracts username and skillID from a UserSkill entity ID
// Returns username, skillID, or empty strings if invalid format
func ParseUserSkillEntityID(entityID string) (username, skillID string) {
	parts := strings.Split(entityID, "#")
	if len(parts) == 3 && parts[0] == "USERSKILL" {
		return parts[1], parts[2]
	}
	return "", ""
}

// ParseMasterSkillEntityID extracts the skillID from a MasterSkill entity ID
// Returns the skillID or empty string if invalid format
func ParseMasterSkillEntityID(entityID string) string {
	parts := strings.Split(entityID, "#")
	if len(parts) == 2 && parts[0] == "SKILL" {
		return parts[1]
	}
	return ""
}
