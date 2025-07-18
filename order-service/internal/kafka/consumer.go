package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"order-service/internal/data_base"
	"order-service/internal/models"
	"strings"

	"github.com/segmentio/kafka-go"
)

func StartConsumer(ctx context.Context, db *data_base.Postgres) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{"kafka:9092"},
		Topic:    "orders-topic",
		GroupID:  "order-service",
		MinBytes: 10e3,
		MaxBytes: 10e6,
	})
	defer reader.Close()

	log.Println("Kafka consumer started")

	for {
		msg, err := reader.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return // graceful shutdown
			}
			log.Printf("[ERROR] Reading message: %v\n", err)
			continue
		}

		// Логируем сырое сообщение для отладки
		log.Printf("[DEBUG] Received message: %s\n", string(msg.Value))

		var order models.Order
		if err := validateAndParse(msg.Value, &order); err != nil {
			log.Printf("[ERROR] Invalid message: %v\nRaw: %s\n", err, string(msg.Value))
			continue
		}

		if err := db.SaveOrder(ctx, order); err != nil {
			log.Printf("[ERROR] DB save failed for order %s: %v\n", order.OrderUID, err)
			continue
		}

		log.Printf("[INFO] Successfully processed order: %s\n", order.OrderUID)
	}
}

// validateAndParse проверяет и парсит JSON сообщение
func validateAndParse(data []byte, order *models.Order) error {
	// 1. Проверка валидности JSON
	if !json.Valid(data) {
		return fmt.Errorf("invalid JSON format")
	}

	// 2. Попытка парсинга
	decoder := json.NewDecoder(strings.NewReader(string(data)))
	decoder.DisallowUnknownFields() // Запрещаем неизвестные поля

	if err := decoder.Decode(order); err != nil {
		// Определяем тип ошибки
		switch {
		case strings.Contains(err.Error(), "unknown field"):
			return fmt.Errorf("unknown field in JSON")
		case strings.Contains(err.Error(), "missing required field"):
			return fmt.Errorf("missing required field")
		default:
			return fmt.Errorf("parsing error: %v", err)
		}
	}

	// 3. Валидация обязательных полей
	if order.OrderUID == "" {
		return fmt.Errorf("empty order_uid")
	}
	if len(order.Items) == 0 {
		return fmt.Errorf("order must contain at least one item")
	}

	return nil
}
