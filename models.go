package main

import (
	"time"
)

type Order struct {
	ID          int     `json:"id" db:"id"`
	CustomerID  string  `json:"customer_id" db:"customer_id"`
	ProductCode string  `json:"product_code" db:"product_code"`
	Quantity    int     `json:"quantity" db:"quantity"`
	Price       float64 `json:"price" db:"price"`
	TotalAmount float64 `json:"total_amount" db:"total_amount"`
	Status      string  `json:"status" db:"status"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type CreateOrderRequest struct {
	CustomerID  string `json:"customer_id" binding:"required"`
	ProductCode string `json:"product_code" binding:"required"`
	Quantity    int    `json:"quantity" binding:"required,min=1"`
}

type UpdateOrderRequest struct {
	CustomerID  string `json:"customer_id"`
	ProductCode string `json:"product_code"`
	Quantity    int    `json:"quantity"`
}

type UpdateOrderStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=pending processing shipped delivered cancelled"`
}

type APIResponse struct {
	Status string        `json:"status"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}