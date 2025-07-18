package data_base

import (
	"context"
	"errors"
	"fmt"
	"order-service/internal/models"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Postgres struct {
	pool *pgxpool.Pool
}

var ErrOrderNotFound = errors.New("order not found")

func NewPostgres(ctx context.Context, connString string) (*Postgres, error) {
	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to DB: %v", err)
	}

	err = pool.Ping(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to ping DB: %v", err)
	}
	return &Postgres{pool: pool}, nil
}

func (p *Postgres) Close(ctx context.Context) error {
	if p.pool != nil {
		p.pool.Close()
	}
	return nil
}

func (p *Postgres) SaveOrder(ctx context.Context, order models.Order) error {
	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
	INSERT INTO orders (
	order_uid, track_number, entry, locale,
	 internal_signature, customer_id, delivery_service,
	 shardkey, sm_id, date_created, oof_shard
	 ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		order.OrderUID, order.TrackNumber, order.Entry, order.Locale,
		order.InternalSignature, order.CustomerID, order.DeliveryService,
		order.ShardKey, order.SmID, order.DateCreated, order.OofShard,
	)
	if err != nil {
		return fmt.Errorf("insert order failed: %w", err)
	}

	_, err = tx.Exec(ctx, `
	INSERT INTO deliveries (order_uid, name, phone, zip, city, address, region,
	email) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		order.OrderUID, order.Delivery.Name, order.Delivery.Phone, order.Delivery.Zip,
		order.Delivery.City, order.Delivery.Address, order.Delivery.Region, order.Delivery.Email,
	)
	if err != nil {
		return fmt.Errorf("insert delivery failed %w", err)
	}

	_, err = tx.Exec(ctx, `
	INSERT INTO payments (order_uid, transaction, request_id, currency, provider,
	amount, payment_dt, bank, delivery_cost, goods_total, custom_fee
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		order.OrderUID, order.Payment.Transaction, order.Payment.RequestID,
		order.Payment.Currency, order.Payment.Provider,
		order.Payment.Amount, order.Payment.PaymentDt, order.Payment.Bank,
		order.Payment.DeliveryCost, order.Payment.GoodsTotal, order.Payment.CustomFee,
	)
	if err != nil {
		return fmt.Errorf("insert payment failed %w", err)
	}

	for _, item := range order.Items {
		_, err = tx.Exec(ctx, `
		INSERT INTO items (order_uid, chrt_id, track_number, price, rid, name, sale,
		size, total_price, nm_id, brand, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`,
			order.OrderUID, item.ChrtID, item.TrackNumber, item.Price,
			item.Rid, item.Name, item.Sale, item.Size,
			item.TotalPrice, item.NmID, item.Brand, item.Status)
		if err != nil {
			return fmt.Errorf("insert item failed %w", err)
		}
	}

	return tx.Commit(ctx)
}

func (p *Postgres) GetOrderByUID(ctx context.Context, orderUID string) (*models.Order, error) {
	order := &models.Order{}

	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		}
	}()

	// Основные данные заказа
	err = tx.QueryRow(ctx, `
        SELECT order_uid, track_number, entry, locale, internal_signature,
               customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard
        FROM orders WHERE order_uid = $1`, orderUID).
		Scan(&order.OrderUID, &order.TrackNumber, &order.Entry, &order.Locale,
			&order.InternalSignature, &order.CustomerID, &order.DeliveryService,
			&order.ShardKey, &order.SmID, &order.DateCreated, &order.OofShard)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrOrderNotFound
		}
		return nil, fmt.Errorf("failed to scan order: %v", err)
	}

	// Доставка
	err = tx.QueryRow(ctx, `
        SELECT name, phone, zip, city, address, region, email
        FROM deliveries WHERE order_uid = $1`, orderUID).
		Scan(&order.Delivery.Name, &order.Delivery.Phone, &order.Delivery.Zip,
			&order.Delivery.City, &order.Delivery.Address, &order.Delivery.Region,
			&order.Delivery.Email)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("failed to scan delivery: %v", err)
	}

	// Платеж
	err = tx.QueryRow(ctx, `
        SELECT transaction, request_id, currency, provider, amount, payment_dt,
               bank, delivery_cost, goods_total, custom_fee
        FROM payments WHERE order_uid = $1`, orderUID).
		Scan(&order.Payment.Transaction, &order.Payment.RequestID, &order.Payment.Currency,
			&order.Payment.Provider, &order.Payment.Amount, &order.Payment.PaymentDt,
			&order.Payment.Bank, &order.Payment.DeliveryCost, &order.Payment.GoodsTotal,
			&order.Payment.CustomFee)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("failed to scan payment: %v", err)
	}

	// Товары
	rows, err := tx.Query(ctx, `
        SELECT chrt_id, track_number, price, rid, name, sale, size,
               total_price, nm_id, brand, status
        FROM items WHERE order_uid = $1`, orderUID)
	if err != nil {
		return nil, fmt.Errorf("failed to query items: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var item models.Item
		if err := rows.Scan(&item.ChrtID, &item.TrackNumber, &item.Price, &item.Rid,
			&item.Name, &item.Sale, &item.Size, &item.TotalPrice, &item.NmID,
			&item.Brand, &item.Status); err != nil {
			return nil, fmt.Errorf("failed to scan item: %v", err)
		}
		order.Items = append(order.Items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("items rows error: %v", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %v", err)
	}

	return order, nil
}
