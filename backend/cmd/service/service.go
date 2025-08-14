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
	// Чтение кофигурационных файлов
	cfg := config.MustLoad()
	// Получение эземпляра базы данных
	storage, err := postgres.NewInstance(cfg.PostgresConnection.DataBasePath())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	_ = storage

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
			fmt.Println(string(msg.Value))
		}
	}()

	// Закрытие сервиса
	exit := make(chan os.Signal, 1)
	// Подписка канала на получение сигналов:
	signal.Notify(exit, os.Interrupt, syscall.SIGTERM)
	<-exit
	fmt.Println("Server was shut down")
}
