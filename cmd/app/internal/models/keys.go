package models

import (
	"fmt"
	"strings"
)

// BuildUserEntityID constructs the entity_id for a User
// Format: USER#<username>
func BuildUserEntityID(username string) string {
	return fmt.Sprintf("USER#%s", strings.ToLower(username))
}

// BuildMasterSkillEntityID constructs the entity_id for a Master Skill
// Format: SKILL#<skill_id>
func BuildMasterSkillEntityID(skillID string) string {
	return fmt.Sprintf("SKILL#%s", skillID)
}

// BuildUserSkillEntityID constructs the entity_id for a User Skill
// Format: USERSKILL#<username>#<skill_id>
func BuildUserSkillEntityID(username, skillID string) string {
	return fmt.Sprintf("USERSKILL#%s#%s", username, skillID)
}
