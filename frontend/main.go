package main

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/segmentio/kafka-go"
)

var (
	messageChan = make(chan string, 100)
)

type KafkaProducer struct {
	writers map[string]*kafka.Writer
}

type PageData struct {
	Message          string
	Status           string
	ResponseMessages []string
}

// HTML шаблон
const htmlTemplate = ``

var (
	responseMessages = make([]string, 0)
)

func NewKafkaProducer(brokers []string, topics []string) *KafkaProducer {
	writers := make(map[string]*kafka.Writer)

	for _, topic := range topics {
		writer := &kafka.Writer{
			Addr:  kafka.TCP(brokers...),
			Topic: topic,
		}
		writers[topic] = writer
	}

	return &KafkaProducer{
		writers: writers,
	}
}

// SendMessage отправляет сырые байты, без обёрток и повторной сериализации.
// Для json_data (опционально) валидирует, что message — валидный JSON.
func (kp *KafkaProducer) SendMessage(ctx context.Context, topic, message string) error {
	writer, exists := kp.writers[topic]
	if !exists {
		return fmt.Errorf("writer for topic %s not found", topic)
	}

	raw := []byte(message)

	// Опциональная валидация для json_data: убедиться, что это валидный JSON
	if topic == "json_data" {
		var tmp json.RawMessage
		if err := json.Unmarshal(raw, &tmp); err != nil {
			return fmt.Errorf("invalid JSON for topic %s: %w", topic, err)
		}
	}

	// Логируем сырые байты для отладки (усекаем до 1KB)
	logPreview := message
	if len(logPreview) > 1024 {
		logPreview = logPreview[:1024] + "...(truncated)"
	}
	log.Printf("Sending raw payload to topic %s: %s", topic, logPreview)

	// Создаем Kafka сообщение
	kafkaMessage := kafka.Message{
		Key:   []byte(fmt.Sprintf("%d", time.Now().Unix())),
		Value: raw,
		Time:  time.Now(),
	}

	// Отправляем сообщение
	if err := writer.WriteMessages(ctx, kafkaMessage); err != nil {
		return fmt.Errorf("failed to write message to topic %s: %w", topic, err)
	}

	log.Printf("Message sent to topic %s successfully", topic)
	return nil
}

func (kp *KafkaProducer) Close() error {
	var lastErr error
	for topic, writer := range kp.writers {
		if err := writer.Close(); err != nil {
			log.Printf("Error closing writer for topic %s: %v", topic, err)
			lastErr = err
		}
	}
	return lastErr
}

// Функция для проверки доступности топиков
func checkTopics(brokers []string, topics []string) error {
	conn, err := kafka.Dial("tcp", brokers[0])
	if err != nil {
		return fmt.Errorf("failed to dial kafka: %w", err)
	}
	defer conn.Close()

	partitions, err := conn.ReadPartitions()
	if err != nil {
		return fmt.Errorf("failed to read partitions: %w", err)
	}

	existingTopics := make(map[string]bool)
	for _, partition := range partitions {
		existingTopics[partition.Topic] = true
	}

	for _, topic := range topics {
		if !existingTopics[topic] {
			log.Printf("Warning: Topic '%s' does not exist in Kafka", topic)
		} else {
			log.Printf("Topic '%s' is available", topic)
		}
	}

	return nil
}

func startResponseConsumer(brokers []string, topic string) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: brokers,
		GroupID: "response-consumer-group",
		Topic:   topic,
	})

	go func() {
		defer reader.Close()

		for {
			msg, err := reader.ReadMessage(context.Background())
			if err != nil {
				log.Printf("Failed to read message from %s: %v", topic, err)
				continue
			}

			text := string(msg.Value)
			log.Printf("Received response message: %s", text)

			// Отправляем в SSE канал
			select {
			case messageChan <- text:
			default:
				log.Println("Message channel is full, dropping message")
			}
		}
	}()
}

func main() {
	brokers := []string{"localhost:9092"}
	topics := []string{"json_data", "order_id", "order_response"}

	// Проверяем доступность топиков
	if err := checkTopics(brokers, topics); err != nil {
		log.Printf("Warning: Could not verify topics: %v", err)
	}

	// Создаем Kafka producer
	kafkaProducer := NewKafkaProducer(brokers, topics)
	defer func() {
		if err := kafkaProducer.Close(); err != nil {
			log.Printf("Error closing Kafka producer: %v", err)
		}
	}()

	// Запускаем consumer, который слушает order_response и сохраняет сообщения
	startResponseConsumer(brokers, "order_response")

	// Парсим HTML шаблон
	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		log.Fatalf("Failed to parse template: %v", err)
	}

	// Обработчик главной страницы
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		data := PageData{
			ResponseMessages: append([]string(nil), responseMessages...), // копия среза для безопасности
		}

		if r.Method == "POST" {
			message := r.FormValue("message")
			topic := r.FormValue("topic")

			if message == "" {
				data.Status = "error"
				data.Message = "Сообщение не может быть пустым"
			} else if topic != "json_data" && topic != "order_id" {
				data.Status = "error"
				data.Message = "Неверный топик"
			} else {
				// Создаем контекст с таймаутом
				ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				defer cancel()

				if err := kafkaProducer.SendMessage(ctx, topic, message); err != nil {
					data.Status = "error"
					data.Message = fmt.Sprintf("Ошибка отправки: %v", err)
					log.Printf("Error sending message: %v", err)
				} else {
					data.Status = fmt.Sprintf("Сообщение успешно отправлено в топик '%s'", topic)
					data.Message = message // Сохраняем отправленное сообщение для формы
				}
			}
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := tmpl.Execute(w, data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	// Добавляем health check endpoint
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"status":    "healthy",
			"timestamp": time.Now().Format(time.RFC3339),
		})
	})

	http.HandleFunc("/events", func(w http.ResponseWriter, r *http.Request) {
		// Настройка SSE headers
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		// Создаем flusher для принудительной отправки данных
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
			return
		}

		// Отправляем сообщения на сайт
		for {
			select {
			case message := <-messageChan:
				fmt.Fprintf(w, "data: %s\n\n", message)
				flusher.Flush()
			case <-r.Context().Done():
				return
			}
		}
	})

	// Запускаем сервер
	port := ":8080"
	log.Printf("Server starting on http://localhost%s", port)
	log.Printf("Health check available at http://localhost%s/health", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("Failede to launch server; Error: %v", err)
	}
}
