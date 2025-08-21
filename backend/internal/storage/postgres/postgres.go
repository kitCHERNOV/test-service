package postgres

import (
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"test/internal/models"
	"time"
)

type Storage struct {
	db *gorm.DB
}

func NewInstance(storagePath string) (*Storage, error) {
	const op = "storage.postgres.New"

	//storagePath = "postgres://postgres:1234@localhost:5432/service?sslmode=disable"
	fmt.Printf("[DEBUG] %s: Attempting connection with DSN: %s\n", op, storagePath)
	var db *gorm.DB
	var err error
	db, err = gorm.Open(postgres.Open(storagePath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info), // Включаем логирование для диагностики
	})

	if err != nil {
		log.Fatalf("Loc: %s, Err: %v", op, err)
	}
	// Параметры соединения
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("%s: unwrap *sql.DB: %w", op, err)
	}
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(25)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("%s: unwrap *sql.DB: %w", op, err)
	}

	return &Storage{
		db: db,
	}, nil
}

// Функция автоматической миграции
func (s *Storage) AutoMigrate() error {
	return s.db.AutoMigrate(
		&models.Order{},
		&models.Delivery{},
		&models.Payment{},
		&models.Item{},
	)
}

// Функция загрузки полученных данных в БД
func (s *Storage) NewDataLoad(order *models.Order) error {
	const op = "storage.postgres.NewDataLoad"

	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(order).Error; err != nil {
			return err
		}
		return nil
	})
	return err
}

func (s *Storage) GetOrderByUID(orderUID string) (*models.Order, error) {
	const op = "storage.postgres.GetOrderByUID"
	var order models.Order

	res := s.db.Preload("Delivery").
		Preload("Payment").
		Preload("Items").
		Where("order_uid = ?", orderUID).First(&order)

	if res.Error != nil {
		return nil, res.Error
	}

	return &order, nil
}

// Функция для извлечения максимум n-го числа данных
func (s *Storage) GetDataToRestoreCache(uids []string) (map[string]models.Order, error) {
	const op = "storage.postgres.GetDataToRestoreCache"
	// Возвращаемая маппа заказов
	orders := make(map[string]models.Order)

	if len(uids) == 0 {
		return orders, nil
	}
	// Получаем данные в список
	var ordersSlice []models.Order
	result := s.db.Preload("Delivery").
		Preload("Payment").
		Preload("Items").
		Where("order_uid IN (?)", uids).
		Find(&ordersSlice)

	if result.Error != nil {
		return nil, fmt.Errorf("Loc:%s: Err:%v", op, result.Error)
	}
	// Преобразуем slice в map
	for _, order := range ordersSlice {
		orders[order.OrderUID] = order
	}

	return orders, nil
}
