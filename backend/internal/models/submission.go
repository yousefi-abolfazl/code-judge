package models

import (
	"time"
)

type SubmissionStatus string

const (
	StatusPending    SubmissionStatus = "pending"
	StatusProcessing SubmissionStatus = "processing"
	StatusAccepted   SubmissionStatus = "accepted"
	StatusRejected   SubmissionStatus = "rejected"
	StatusError      SubmissionStatus = "error"
)

type SubmissionResult string

const (
	ResultOK           SubmissionResult = "OK"
	ResultCompileError SubmissionResult = "Compile Error"
	ResultWrongAnswer  SubmissionResult = "Wrong Answer"
	ResultMemoryLimit  SubmissionResult = "Memory Limit"
	ResultTimeLimit    SubmissionResult = "Time Limit"
	ResultRuntimeError SubmissionResult = "Runtime Error"
)

type Submission struct {
	ID            uint             `gorm:"primaryKey" json:"id"`
	UserID        uint             `json:"user_id"`
	ProblemID     uint             `json:"problem_id"`
	Code          string           `gorm:"type:text;not null" json:"code"`
	Status        SubmissionStatus `gorm:"default:pending;not null" json:"status"`
	Result        SubmissionResult `json:"result"`
	ExecutionTime int              `json:"execution_time"` // میلی‌ثانیه
	MemoryUsed    int              `json:"memory_used"`    // مگابایت
	ErrorMessage  string           `gorm:"type:text" json:"error_message"`
	CreatedAt     time.Time        `json:"created_at"`
	UpdatedAt     time.Time        `json:"updated_at"`
}
