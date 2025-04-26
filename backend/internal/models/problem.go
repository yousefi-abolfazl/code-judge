package models

import (
	"time"
)

type ProblemStatus string

const (
	ProblemDraft     ProblemStatus = "draft"
	ProblemPublished ProblemStatus = "published"
)

type Problem struct {
	ID          uint          `gorm:"primaryKey" json:"id"`
	Title       string        `gorm:"size:255;not null" json:"title"`
	Statement   string        `gorm:"type:text;not null" json:"statement"`
	TimeLimit   int           `gorm:"not null" json:"time_limit"`   // میلی‌ثانیه
	MemoryLimit int           `gorm:"not null" json:"memory_limit"` // مگابایت
	Input       string        `gorm:"type:text;not null" json:"input"`
	Output      string        `gorm:"type:text;not null" json:"output"`
	Status      ProblemStatus `gorm:"default:draft;not null" json:"status"`
	PublishDate *time.Time    `json:"publish_date"`
	OwnerID     uint          `json:"owner_id"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
}
