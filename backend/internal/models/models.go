package models

import (
	"net/http"
	"time"

	"gorm.io/gorm"
	_ "gorm.io/gorm"
)

// Может быть будет эффективно обращаться через интерфейс
type OrderService interface {
	Load(w http.ResponseWriter, r http.Request)
}

type Order struct {
	OrderUID    string `gorm:"primaryKey" json:"order_uid"`
	TrackNumber string `gorm:"size:255;not null" json:"track_number"`
	Entry       string `gorm:"" json:"entry"`

	CreatedAt time.Time      `gorm:"autoCreateTime"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}
