package broker

import (
	"encoding/json"
	"errors"
	"fmt"
	"test/internal/models"
)

func UnmarshalingOrderDataMessages(message []byte) (*models.Order, error) {
	const op = "handlers.kafka.UnmarshalingOrderDataMessages"
	var order models.Order

	err := json.Unmarshal(message, &order)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Loc:%s; Err:%s", op, err.Error()))
	}

	return &order, nil
}

// Имплементация продюсера

func UnmarshalingOrderId(message []byte) (*models.OrderID, error) {
	const op = "handlers.kafka.UnmarshalingOrderId"

	var order models.OrderID
	err := json.Unmarshal(message, &order)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Loc:%s; Err:%s", op, err.Error()))
	}
	return &order, nil
}

func MarshalingOrderDataMessages(order *models.Order) ([]byte, error) {
	msg, err := json.Marshal(order)
	if err != nil {
		return nil, err
	}
	return msg, nil
}
