package models

import (
	_ "gorm.io/gorm"
	"net/http"
)

// Может быть будет эффективно обращаться через интерфейс
type OrderService interface {
	Load(w http.ResponseWriter, r http.Request)
}

type Order struct {
	OrderUID    string `gorm:"order_uid" json:"order_uid"`
	TrackNumber string `gorm:"track_number" json:"track_number"`
	Entry       string `gorm:"entry" json:"entry"`
}

func (o *Order) Load(w http.ResponseWriter, r http.Request) {
	var gone string = "12"
	var done string = "13"
	_ = gone
	_ = done

}
