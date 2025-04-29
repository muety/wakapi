package models

import "time"

// UserReportSent represents a record indicating that a report has been sent
// for a specific user on a specific date. Used to prevent duplicate emails
// in a distributed environment. This is just duct-tape - until I actually can think of something better.
// reports are to be sent within the next 5 hours. I need to patch this asap
type UserReportSent struct {
	ID         uint      `gorm:"primaryKey"`
	UserID     string    `gorm:"not null;uniqueIndex:idx_user_report_date"`
	ReportDate time.Time `gorm:"not null;uniqueIndex:idx_user_report_date"`
	SentAt     time.Time `gorm:"not null"`

	User User `gorm:"foreignKey:UserID"`
}

// TableName specifies the table name for GORM
func (UserReportSent) TableName() string {
	return "user_reports_sent"
}
