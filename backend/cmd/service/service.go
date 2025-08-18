package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"test/internal/config"
	"test/internal/handlers/broker"
	"test/internal/storage/cache"
	"test/internal/storage/postgres"

	"github.com/segmentio/kafka-go"
)

func ensureTopic(broker string, cfg kafka.TopicConfig) error {
	conn, err := kafka.Dial("tcp", broker)
	if err != nil {
		return err
	}
	defer conn.Close()

	controller, err := conn.Controller()
	if err != nil {
		return err
	}

	ctrlAddr := net.JoinHostPort(controller.Host, strconv.Itoa(controller.Port))
	cconn, err := kafka.Dial("tcp", ctrlAddr)
	if err != nil {
		return err
	}
	defer cconn.Close()

	return cconn.CreateTopics(cfg)
}

func main() {
	var err error
	var ctx context.Context = context.Background()
	// Чтение кофигурационных файлов
	cfg := config.MustLoad()
	// Получение эземпляра базы данных
	storage, err := postgres.NewInstance(cfg.PostgresConnection.DataBasePath())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Автоматическая миграция
	err = storage.AutoMigrate()
	if err != nil {
		log.Fatalf("Failed to auto migrate: %v", err)
	}
	// Инициализация кэша
	cacheInstance := cache.InitCache(cfg.CacheParams)

	// Организация топиков кафки
	err = ensureTopic(cfg.Broker, kafka.TopicConfig{
		Topic:             "order_id",
		NumPartitions:     3,
		ReplicationFactor: 1,
	})
	if err != nil {
		log.Fatalf("Failed to create topic: %v", err)
	}

	// Подписка на топики:
	// 1. топик order_id
	go func() {
		readerOrderId := kafka.NewReader(kafka.ReaderConfig{
			// TODO: Сохранять поле Broker в config как список string
			Brokers: []string{cfg.Broker},
			Topic:   "order_id",
		})
		responseWriterOrderID := kafka.Writer{
			Addr:  kafka.TCP(cfg.Broker),
			Topic: "order_response",
		}
		defer func() {
			err = readerOrderId.Close()
			if err != nil {
				log.Fatalf("Failed to close id reader: %v", err)
			}
		}()
		fmt.Println("Log: Servise is ready to get order_id messages")
		for {
			msg, err := readerOrderId.ReadMessage(context.Background())
			if err != nil {
				log.Fatalf("Failed to read message: %v", err)
			}

			// Получим интересуемый ID для поиска данных по заказу
			orderIdStruct, err := broker.UnmarshalingOrderId(msg.Value)
			if err != nil {
				log.Fatalf("Failed to unmarshal id message: %v", err)
			}
			var takenData []byte
			// Постараемся получить сообщение из кэша
			if order, ok := cacheInstance[orderIdStruct.OrderUID]; ok {
				takenData, err = broker.MarshalingOrderDataMessages(&order)
				if err != nil {
					log.Fatalf("Failed to marshal order: %v", err)
				}
			}
			// Если заказ не был найден в кэше, то найдем его в Базе данных
			// TODO: Реализовать поиск данных по ключу
			response := kafka.Message{
				Value: takenData,
			}
			err = responseWriterOrderID.WriteMessages(ctx, response)
			if err != nil {
				log.Fatalf("Failed to write message: %v", err)
			}
			fmt.Println(string(msg.Value))
		}
	}()
	// 2. топик json_data
	go func() {
		readerOrderJson := kafka.NewReader(kafka.ReaderConfig{
			Brokers: []string{cfg.Broker},
			Topic:   "json_data",
		})
		defer func() {
			err = readerOrderJson.Close()
			if err != nil {
				log.Fatalf("Failed to close json reader: %v", err)
			}
		}()
		fmt.Println("Log: Servise is ready to get json_data messages")
		for {
			msg, err := readerOrderJson.ReadMessage(context.Background())
			if err != nil {
				log.Fatalf("Failed to read message: %v", err)
			}
			order, err := broker.UnmarshalingOrderDataMessages(msg.Value)
			if err != nil {
				log.Fatalf("Failed to unmarshal json data: %v", err)
			}
			// Запись в бд
			err = storage.NewDataLoad(order)
			if err != nil {
				log.Fatalf("Failed to store order: %v", err)
			}
			// Сохранение в кэш
			cacheInstance.CacheItemAdd(order.OrderUID, *order)
		}
	}()

	// Закрытие сервиса
	exit := make(chan os.Signal, 1)
	// Подписка канала на получение сигналов:
	signal.Notify(exit, os.Interrupt, syscall.SIGTERM)
	<-exit
	fmt.Println("Server was shut down")
}
