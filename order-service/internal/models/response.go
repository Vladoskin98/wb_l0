package models

import "time"

type OrderResponse struct {
	Order      *Order `json:"order"`
	Source     string `json:"source"`
	DurationMs int64  `json:"duration_ms"`
	Cached     bool   `json:"cached"`
}

func NewOrderResponse(order *Order, source string, duration time.Duration) *OrderResponse {
	return &OrderResponse{
		Order:      order,
		Source:     source,
		DurationMs: duration.Milliseconds(),
		Cached:     source == "cache",
	}
}
