package models

import (
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	MobileNumber string  `gorm:"uniqueIndex;not null" json:"mobile_number"`
	Name         string  `gorm:"not null" json:"name"`
	Email        *string `json:"email,omitempty"`
	Role         string  `gorm:"not null" json:"role"`
	ProfilePhoto *string `json:"profile_photo,omitempty"`
}

type Shop struct {
	gorm.Model
	ShopID       string  `gorm:"uniqueIndex;not null" json:"shop_id"`
	ShopName     string  `gorm:"not null" json:"shop_name"`
	ShopUsername string  `gorm:"not null" json:"shop_username"`
	Bio          *string `json:"bio,omitempty"`
	ProductType  string  `gorm:"not null" json:"product_type"`
	Location     *string `json:"location,omitempty"`
	ShopPhoto    *string `json:"shop_photo,omitempty"`
	UserID       uint    `gorm:"not null;uniqueIndex" json:"userid"`                     // Ensures one shop per user
	User         User    `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"` // Foreign key constraint
}
