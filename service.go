package main

import (
	"database/sql"
	"fmt"
)

type OrderService struct {
	db *sql.DB
}

func NewOrderService(db *sql.DB) *OrderService {
	return &OrderService{db: db}
}

func (s *OrderService) CreateOrder(req CreateOrderRequest) (*Order, error) {
	// For now, we'll set a default price - in a real scenario, you'd lookup the product price
	// from a products table or external service using the product_code
	defaultPrice := 10.00 // You can modify this or add product lookup logic
	totalAmount := float64(req.Quantity) * defaultPrice
	
	var order Order
	query := `
		INSERT INTO orders (customer_id, product_code, quantity) 
		VALUES ($1, $2, $3) 
		RETURNING id, customer_id, product_code, quantity, created_at, updated_at
	`
	
	err := s.db.QueryRow(query, req.CustomerID, req.ProductCode, req.Quantity).Scan(
		&order.ID, &order.CustomerID, &order.ProductCode, &order.Quantity, 
		&order.CreatedAt, &order.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	order.Price = defaultPrice
	order.TotalAmount = totalAmount
	order.Status = "pending"

	return &order, nil
}

func (s *OrderService) GetOrder(id int) (*Order, error) {
	var order Order
	query := `SELECT id, customer_id, product_code, quantity, price, total_amount, status, created_at, updated_at FROM orders WHERE id = $1`
	
	err := s.db.QueryRow(query, id).Scan(
		&order.ID, &order.CustomerID, &order.ProductCode, &order.Quantity,
		&order.Price, &order.TotalAmount, &order.Status, &order.CreatedAt, &order.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("order not found")
		}
		return nil, err
	}

	return &order, nil
}

func (s *OrderService) GetOrders(limit, offset int, status string) ([]Order, error) {
	if limit == 0 {
		limit = 10
	}

	var query string
	var args []interface{}

	if status != "" {
		query = `SELECT id, customer_id, product_code, quantity, price, total_amount, status, created_at, updated_at FROM orders WHERE status = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`
		args = []interface{}{status, limit, offset}
	} else {
		query = `SELECT id, customer_id, product_code, quantity, price, total_amount, status, created_at, updated_at FROM orders ORDER BY created_at DESC LIMIT $1 OFFSET $2`
		args = []interface{}{limit, offset}
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []Order
	for rows.Next() {
		var order Order
		err := rows.Scan(&order.ID, &order.CustomerID, &order.ProductCode, &order.Quantity,
			&order.Price, &order.TotalAmount, &order.Status, &order.CreatedAt, &order.UpdatedAt)
		if err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}

	return orders, nil
}

func (s *OrderService) UpdateOrder(id int, req UpdateOrderRequest) (*Order, error) {
	// Check if order exists
	existingOrder, err := s.GetOrder(id)
	if err != nil {
		return nil, err
	}

	// Calculate new total if quantity changed
	quantity := existingOrder.Quantity
	price := existingOrder.Price
	
	if req.Quantity > 0 {
		quantity = req.Quantity
	}
	
	totalAmount := float64(quantity) * price

	query := `
		UPDATE orders 
		SET customer_id = COALESCE(NULLIF($1, ''), customer_id),
		    product_code = COALESCE(NULLIF($2, ''), product_code),
		    quantity = CASE WHEN $3 > 0 THEN $3 ELSE quantity END,
		    total_amount = $4
		WHERE id = $5
		RETURNING id, customer_id, product_code, quantity, price, total_amount, status, created_at, updated_at
	`
	
	var order Order
	err = s.db.QueryRow(query, req.CustomerID, req.ProductCode, req.Quantity, totalAmount, id).Scan(
		&order.ID, &order.CustomerID, &order.ProductCode, &order.Quantity,
		&order.Price, &order.TotalAmount, &order.Status, &order.CreatedAt, &order.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &order, nil
}

func (s *OrderService) UpdateOrderStatus(id int, status string) (*Order, error) {
	// Check if order exists
	_, err := s.GetOrder(id)
	if err != nil {
		return nil, err
	}

	query := `
		UPDATE orders 
		SET status = $1
		WHERE id = $2
		RETURNING id, customer_id, product_code, quantity, price, total_amount, status, created_at, updated_at
	`
	
	var order Order
	err = s.db.QueryRow(query, status, id).Scan(
		&order.ID, &order.CustomerID, &order.ProductCode, &order.Quantity,
		&order.Price, &order.TotalAmount, &order.Status, &order.CreatedAt, &order.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &order, nil
}

func (s *OrderService) DeleteOrder(id int) error {
	query := `DELETE FROM orders WHERE id = $1`
	result, err := s.db.Exec(query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("order not found")
	}

	return nil
}