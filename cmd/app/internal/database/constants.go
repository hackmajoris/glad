package database

const (
	// TableName is the single table for all entities
	TableName = "glad-entities"

	GSIByUser        = "ByUser"
	GSISkillsByLevel = "SkillsByLevel"
	GSIBySkillID     = "BySkillID"
	GSIByEntityType  = "ByEntityType"
)
