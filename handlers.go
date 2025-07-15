package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	amqp "github.com/rabbitmq/amqp091-go"
)

type OrderHandler struct {
	orderService *OrderService
	rabbitMQ     *amqp.Connection
    channel      *amqp.Channel
}

func NewOrderHandler(orderService *OrderService) (*OrderHandler, error) {
    // Connect to RabbitMQ
	var rabbitMQURL = os.Getenv("RABBITMQ_URL")
    conn, err := amqp.Dial(rabbitMQURL)
    if err != nil {
        return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
    }

    ch, err := conn.Channel()
    if err != nil {
        conn.Close() // Clean up connection on error
        return nil, fmt.Errorf("failed to open channel: %w", err)
    }

    // Declare exchange (optional, depends on your setup)
    err = ch.ExchangeDeclare(
        "orders",   // name
        "topic",    // type
        true,       // durable
        false,      // auto-deleted
        false,      // internal
        false,      // no-wait
        nil,        // arguments
    )
    if err != nil {
        ch.Close()
        conn.Close()
        return nil, fmt.Errorf("failed to declare exchange: %w", err)
    }

    return &OrderHandler{
        orderService: orderService,
        rabbitMQ:     conn,
        channel:      ch,
    }, nil
}

func (h *OrderHandler) CreateOrder(c *gin.Context) {
	var req CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Status: "error",
			Message: err.Error(),
		})
		return
	}

	order, err := h.orderService.CreateOrder(req)
	if err != nil {
	log.Printf("Failed to create order: %v", err)
		c.JSON(http.StatusInternalServerError, APIResponse{
			Status: "error",
			Message: "Failed to create order",
		})
		return
	}

	    // Send message to RabbitMQ on success
    if err := h.publishOrderCreated(order); err != nil {
        log.Printf("Failed to publish order created event: %v", err)
        // Note: You might want to decide whether to return error or just log it

        c.JSON(http.StatusInternalServerError, APIResponse{
            Status: "error",
            Message: "Failed to publish order created event",
        })
        return
    }


	c.JSON(http.StatusCreated, APIResponse{
		Status: "success",
		Data: struct {
			Status string `json:"status"`
			ID     int    `json:"id"`
		}{
			Status: order.Status,
			ID:     order.ID,
		},
	})
}

func (h *OrderHandler) publishOrderCreated(order *Order) error {
    // Create message payload
    message := map[string]interface{}{
        "order_id": order.ID,
        "status":   order.Status,
        "created_at": time.Now(),
        "event_type": "order_created",
    }

    messageBody, err := json.Marshal(message)
    if err != nil {
        return fmt.Errorf("failed to marshal message: %w", err)
    }

    // Publish to RabbitMQ
    err = h.channel.Publish(
        "orders",     // exchange
        "order.created", // routing key
        false,        // mandatory
        false,        // immediate
        amqp.Publishing{
            ContentType: "application/json",
            Body:        messageBody,
            Timestamp:   time.Now(),
        },
    )

    if err != nil {
        return fmt.Errorf("failed to publish message: %w", err)
    }

	fmt.Print("Order created and message published to RabbitMQ\n")
    return nil
}