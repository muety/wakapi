package models

// UserOauth stores oauth data for a user
type UserOauth struct {
	ID         string     `json:"id" gorm:"primary_key"`
	UserID     string     `json:"user_id"`
	CreatedAt  CustomTime `json:"created_at" gorm:"default:CURRENT_TIMESTAMP" swaggertype:"string" format:"date" example:"2006-01-02 15:04:05.000"`
	UpdatedAt  CustomTime `json:"updated_at" gorm:"default:CURRENT_TIMESTAMP" swaggertype:"string" format:"date" example:"2006-01-02 15:04:05.000"`
	ProviderID string     `json:"provider_id"`
	Handle     *string    `json:"handle"` //github, gitlab handle
	AvatarUrl  *string    `json:"avatar_url"`
	Email      *string    `json:"email"`
	Provider   string     `json:"provider"`
}
