package models

import (
	"time"
)

type Event struct {
	ID             string `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	Name           string
	TotalQuota     int
	AvailableQuota int
	Version        int
	CreatedAt      time.Time
}

type Booking struct {
	ID        string `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	EventID   string
	UserID    string
	Status    string
	CreatedAt time.Time
}