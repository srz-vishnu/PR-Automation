package domain

import "time"

type Credential struct {
	ID        uint   `gorm:"primaryKey"`
	Username  string `gorm:"unique;not null"`
	Password  string `gorm:"not null"`
	Token     string `gorm:"type:text"`    // store JWT token
	Status    bool   `gorm:"default:true"` // true = active
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Employee struct {
	ID        uint   `gorm:"primaryKey"`
	EmpID     string `gorm:"unique;not null"` // Company employee ID
	Email     string `gorm:"unique;not null"`
	Status    string `gorm:"default:active"`
	Name      string `gorm:"not null"`
	CreatedAt time.Time
	UpdatedAt time.Time

	PullRequests []PullRequest `gorm:"foreignKey:EmployeeID"` // one-to-many relationship
}

type PullRequest struct {
	ID           uint   `gorm:"primaryKey"`
	EmployeeID   uint   `gorm:"not null"` // FK to Employee
	PRLink       string `gorm:"not null"` // Full PR URL
	RepoName     string
	StaffID      string
	PRNumber     int
	Title        string
	Description  string
	Status       string // open, closed, merged, draft
	IsDraft      bool
	LinesAdded   int
	LinesRemoved int
	FilesChanged int
	CommitCount  int
	BranchName   string
	ReviewStatus string
	CreatedAt    *time.Time
	UpdatedAt    *time.Time
	FetchedAt    time.Time `gorm:"autoUpdateTime"` // Last time data was fetched
}

type PRSnapshot struct {
	ID           uint `gorm:"primaryKey"`
	EmployeeID   uint `gorm:"not null"`
	Name         string
	PRID         uint      `gorm:"not null"` // FK to PullRequest
	Date         time.Time `gorm:"not null"` // date (e.g., 2025-07-16)
	Description  string    `gorm:"not null"`
	LinesAdded   int
	LinesRemoved int
	FilesChanged int
	CommitCount  int
	CreatedAt    time.Time
	UpdatedAt    *time.Time
}

type PRReport struct {
	ID         uint   `gorm:"primaryKey"`
	EmpID      string `gorm:"not null"`
	PRLink     string `gorm:"not null"`
	Status     string `gorm:"not null"` // open, merged, etc.
	ReportText string `gorm:"type:text"`
	IsMailSent bool   `gorm:"default:false"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
