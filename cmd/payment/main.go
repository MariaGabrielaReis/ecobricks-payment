package main

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/MariaGabrielaReis/ecobricks-payment/internal/entity"
	"github.com/MariaGabrielaReis/ecobricks-payment/pkg/rabbitmq"
	amqp "github.com/rabbitmq/amqp091-go"
)

// {"order_id": "123", "card_hash": "123", "total": 120.00}
func main() {
	ctx := context.Background()
	ch, err := rabbitmq.OpenChannel()
	if err != nil {
		panic(err)
	}
	defer ch.Close()

	messages := make(chan amqp.Delivery)
	go rabbitmq.Consume(ch, messages, "orders")

	for message := range messages {
		var orderRequest entity.OrderRequest
		err := json.Unmarshal(message.Body, &orderRequest)
		if err != nil {
			slog.Error(err.Error())
			break
		}
		response, err := orderRequest.Process()
		if err != nil {
			slog.Error(err.Error())
			break
		}
		responseJSON, err := json.Marshal(response)
		if err != nil {
			slog.Error(err.Error())
			break
		}
		err = rabbitmq.Publish(ctx, ch, string(responseJSON), "amq.direct")
		if err != nil {
			slog.Error(err.Error())
			break
		}
		message.Ack(false)
		slog.Info("Processed order")
	}
}
