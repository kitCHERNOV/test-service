package main

import (
	"context"
	"fmt"
	"github.com/segmentio/kafka-go"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"test/internal/config"
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
	//storage, err := postgres.New(cfg.PostgresConnection.DataBasePath())
	//if err != nil {
	//	log.Fatalf("Failed to connect to database: %v", err)
	//}

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
		reader := kafka.NewReader(kafka.ReaderConfig{
			// TODO: Сохранять поле Broker в config как список string
			Brokers: []string{cfg.Broker},
			Topic:   "order_id",
		})
		defer func() {
			err = reader.Close()
			if err != nil {
				log.Fatalf("Failed to close reader: %v", err)
			}
		}()
		fmt.Println("Servise is ready to get messages")
		for {
			msg, err := reader.ReadMessage(context.Background())
			if err != nil {
				log.Fatalf("Failed to read message: %v", err)
			}
			switch msg.Topic {
			case "order_id":
				continue
			case "json_data":
				continue
			case "":

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
