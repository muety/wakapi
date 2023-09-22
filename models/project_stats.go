package models

type ProjectStats struct {
	UserId      string
	Project     string
	TopLanguage string
	Count       int64
	First       CustomTime
	Last        CustomTime
}
