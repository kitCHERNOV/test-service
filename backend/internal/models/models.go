package models

import (
	_ "gorm.io/gorm"
	"time"
)

type OrderID struct {
	OrderUID string `json:"order_uid"`
}

type Order struct {
	ID                uint      `gorm:"primarykey"`
	OrderUID          string    `gorm:"uniqueIndex;size:255;not null" json:"order_uid"`
	TrackNumber       string    `gorm:"size:255" json:"track_number"`
	Entry             string    `gorm:"size:25" json:"entry"`
	Locale            string    `gorm:"size:25" json:"locale"`
	InternalSignature string    `gorm:"size:255" json:"internal_signature"`
	CustomerID        string    `gorm:"size:255" json:"customer_id"`
	DeliveryService   string    `gorm:"size:50" json:"delivery_service"`
	ShardKey          string    `gorm:"size:10" json:"shardkey"`
	SmID              int       `json:"sm_id"`
	DateCreated       time.Time `gorm:"not null" json:"date_created"`
	OofShard          string    `gorm:"size:10" json:"oof_shard"`

	// Встроенные таблички
	Delivery Delivery `gorm:"foreignKey:OrderID;constraint:OnDelete:CASCADE" json:"delivery"`
	Payment  Payment  `gorm:"foreignKey:OrderID;constraint:OnDelete:CASCADE" json:"payment"`
	Items    []Item   `gorm:"foreignKey:OrderID;constraint:OnDelete:CASCADE" json:"items"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// табличка доставки;
type Delivery struct {
	ID      uint   `gorm:"primarykey"`
	OrderID uint   `gorm:"not null;index"`
	Name    string `gorm:"size:255;not null" json:"name"`
	Phone   string `gorm:"size:50" json:"phone"`
	Zip     string `gorm:"size:20" json:"zip"`
	City    string `gorm:"size:100" json:"city"`
	Address string `gorm:"size:500" json:"address"`
	Region  string `gorm:"size:100" json:"region"`
	Email   string `gorm:"size:255" json:"email"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// таблица оплаты
type Payment struct {
	ID           uint   `gorm:"primaryKey"`
	OrderID      uint   `gorm:"not null;index"`
	Transaction  string `gorm:"size:255;not null" json:"transaction"`
	RequestID    string `gorm:"size:255" json:"request_id"`
	Currency     string `gorm:"size:10" json:"currency"`
	Provider     string `gorm:"size:100" json:"provider"`
	Amount       int    `json:"amount"`
	PaymentDt    int64  `json:"payment_dt"`
	Bank         string `gorm:"size:100" json:"bank"`
	DeliveryCost int    `json:"delivery_cost"`
	GoodsTotal   int    `json:"goods_total"`
	CustomFee    int    `json:"custom_fee"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Таблица для определения предметов
type Item struct {
	ID          uint   `gorm:"primaryKey"`
	OrderID     uint   `gorm:"not null;index"`
	ChrtID      int    `gorm:"not null" json:"chrt_id"`
	TrackNumber string `gorm:"size:255" json:"track_number"`
	Price       int    `json:"price"`
	RID         string `gorm:"size:255" json:"rid"`
	Name        string `gorm:"size:255" json:"name"`
	Sale        int    `json:"sale"`
	Size        string `gorm:"size:50" json:"size"`
	TotalPrice  int    `json:"total_price"`
	NMID        int    `json:"nm_id"`
	Brand       string `gorm:"size:255" json:"brand"`
	Status      int    `json:"status"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
