package models

import (
	"time"

	"gorm.io/gorm"
)

// OTPStatus represents the status of an OTP
type OTPStatus string

const (
	OTPStatusPending OTPStatus = "pending"
	OTPStatusUsed    OTPStatus = "used"
	OTPStatusExpired OTPStatus = "expired"
)

const (
	InvalidCodeChallengeError = "invalid code challenge"
	InvalidCodeMethodError    = "invalid code challenge method"
)

// PKCE Challenge Method
type ChallengeMethod string

const (
	Plain  ChallengeMethod = "plain"
	SHA256 ChallengeMethod = "S256"
)

func (m ChallengeMethod) String() string {
	return string(m)
}

type OTP struct {
	gorm.Model
	Email           string `gorm:"not null"`
	OTPHash         string `gorm:"not null"`
	CodeChallenge   string `gorm:"not null"`
	ChallengeMethod string `gorm:"not null"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
	ExpiresIn       int64 `gorm:"not null"` // Unix timestamp
	Used            bool
}

func (otp *OTP) IsValid() bool {
	return !otp.Used && time.Now().Unix() < otp.ExpiresIn
}

// Request and response structs
type InitiateOTPRequest struct {
	Email           string `json:"email"`
	CodeChallenge   string `json:"code_challenge"`
	ChallengeMethod string `json:"challenge_method"`
}

type ValidateOTPRequest struct {
	Email        string `json:"email"`
	OTP          string `json:"otp"`
	CodeVerifier string `json:"code_verifier"`
}

type OTPResponse struct {
	Message string `json:"message"`
}

type CreateOTPResponse struct {
	Message string `json:"message"`
	Success bool   `json:"success"`
}

type VerifyOTPResponse struct {
	Message   string `json:"message"`
	Valid     bool   `json:"valid"`
	User      *User  `json:"user"`
	IsNewUser bool   `json:"is_new_user"`
}
