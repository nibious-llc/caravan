package models

import (
	"github.com/google/uuid"
)

type Client struct {
	ClientID    uuid.UUID `gorm:"type:uuid;primaryKey"`
	Secret      string    `gorm:"size:64;primaryKey"`
	Type        string
	DisplayName string
}
