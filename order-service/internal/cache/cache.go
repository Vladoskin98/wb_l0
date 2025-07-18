package cache

import (
	"context"
	"encoding/json"
	"log"
	"order-service/internal/data_base"
	"order-service/internal/models"
	"os"
	"path/filepath"
	"sync"
)

type Cache struct {
	mu       sync.RWMutex
	data     map[string]models.Order
	db       *data_base.Postgres
	keysFile string
}

func New(db *data_base.Postgres, cacheDir string) *Cache {
	keysFile := filepath.Join(cacheDir, "cache_keys.json")

	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		if err := os.MkdirAll(cacheDir, 0755); err != nil {
			log.Printf("Failed to create cache dir: %v\n", err)
		}
	}

	return &Cache{
		data:     make(map[string]models.Order),
		db:       db,
		keysFile: keysFile,
	}
}

func (c *Cache) Get(orderUID string) *models.Order {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if order, exists := c.data[orderUID]; exists {
		return &order
	}
	return nil
}

func (c *Cache) Set(order models.Order) error {
	c.mu.Lock()
	c.data[order.OrderUID] = order
	log.Println(c.data[order.OrderUID])
	c.mu.Unlock()
	log.Println("cached?")
	return c.saveKeys()
}

func (c *Cache) Restore(ctx context.Context) error {
	keys, err := c.loadKeys()
	if err != nil {
		return err
	}

	for _, key := range keys {
		log.Printf("Restoring order %s from db", key)
		order, err := c.db.GetOrderByUID(ctx, key)
		if err != nil {
			log.Printf("Failed to restore order %s: %v", key, err)
			continue
		}

		c.mu.Lock()
		c.data[key] = *order
		c.mu.Unlock()
	}
	log.Printf("%d orders restored to cache", len(keys))
	return nil
}

func (c *Cache) saveKeys() error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	log.Println("save cache -- create slice")

	keys := make([]string, 0, len(c.data))
	for k := range c.data {
		keys = append(keys, k)
	}
	log.Println("save cache -- slice created")
	data, err := json.Marshal(keys)
	if err != nil {
		return err
	}
	log.Println("key saved?")
	return os.WriteFile(c.keysFile, data, 0644)
}

func (c *Cache) loadKeys() ([]string, error) {
	if _, err := os.Stat(c.keysFile); os.IsNotExist(err) {
		return nil, nil
	}

	data, err := os.ReadFile(c.keysFile)
	if err != nil {
		return nil, err
	}

	var keys []string
	if err := json.Unmarshal(data, &keys); err != nil {
		return nil, err
	}

	return keys, nil
}
