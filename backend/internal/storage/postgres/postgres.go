package postgres

import (
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Storage struct {
	db *gorm.DB
}

func testFunc() {
	fmt.Println("Test")
	testPrinting()

}

func testPrinting() {
	if true {
		const TestingConst = "Test"
		fmt.Fprint(os.Stdout, TestingConst)
	}
}

func New(storagePath string) (*Storage, error) {
	const op = "storage.postgres.New"

	db, err := gorm.Open(postgres.Open(storagePath), &gorm.Config{})
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
