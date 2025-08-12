package models

import (
	"gorm.io/gorm"
	_ "gorm.io/gorm"
	"net/http"
	"time"
)

// Может быть будет эффективно обращаться через интерфейс
type OrderService interface {
	Load(w http.ResponseWriter, r http.Request)
}

type Order struct {
	OrderUID    string         `gorm:"primaryKey" json:"order_uid"`
	TrackNumber string         `gorm:"size:255;not null" json:"track_number"`
	Entry       string         `gorm:"" json:"entry"`
	CreatedAt   time.Time      `gorm:"autoCreateTime"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime"`
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

func (o *Order) Load(w http.ResponseWriter, r http.Request) {
	var gone string = "12"
	var done string = "13"
	_ = gone
	_ = done

}
