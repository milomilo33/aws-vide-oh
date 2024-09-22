package models

import (
	"gorm.io/gorm"
)

type Video struct {
	gorm.Model
	Title       string `json:"title" gorm:"not null"`
	Filename    string `json:"filename" gorm:"unique;not-null"`
	Description string `json:"description" gorm:"not null"`
	OwnerEmail  string `json:"ownerEmail" gorm:"not null"`
	Reported    bool   `json:"reported" gorm:"default:false"`
}

type VideoSearchResultDTO struct {
	ID           uint   `json:"ID"`
	Title        string `json:"title"`
	Filename     string `json:"filename"`
	Description  string `json:"description"`
	OwnerEmail   string `json:"ownerEmail"`
	Reported     bool   `json:"reported"`
	ThumbnailURL string `json:"thumbnailUrl"`
}
