package database

import (
	"github.com/hackmajoris/glad-stack/pkg/config"
)

var (
	// TableName is the single table for all entities
	TableName = config.Load().Database.TableName

	GSIBySkill = "BySkill"
)
