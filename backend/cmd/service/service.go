package main

import (
	"log"
	"test/internal/config"
	"test/internal/storage/postgres"
)

func main() {

	// Чтение кофигурационных файлов
	cfg := config.MustLoad()

	// Получение эземпляра базы данных
	storage, err := postgres.New(cfg.PostgresConnection.DataBasePath())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Создание чи роутера
	//r := chi.NewRouter()
	//r.Use(middleware.Logger)
	//r.Get("/", func(w http.ResponseWriter, r *http.Request) {
	//	w.Write([]byte("Hello World"))
	//})
	//http.ListenAndServe(":8080", r)
}
