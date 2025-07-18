package main

import (
	"context"
	"log"
	"order-service/internal/cache"
	"order-service/internal/data_base"
	"order-service/internal/handlers"
	"order-service/internal/kafka"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	time.Sleep(15 * time.Second)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	db, err := data_base.NewPostgres(ctx,
		"postgres://test_admin:admin@postgres:5432/orders_db")
	if err != nil {
		log.Fatalf("Failed to connect to DB: %v\n", err)
	}
	defer db.Close(ctx)

	cache := cache.New(db, "/app/cache_data")
	if err := cache.Restore(ctx); err != nil {
		log.Printf("Failed to restore cache: %v\n", err)
	}

	go kafka.StartConsumer(ctx, db)

	r := gin.Default()
	h := handlers.New(db, cache)

	r.GET("/order/:id", h.GetOrder)
	r.Static("/static", "./static")

	go func() {
		if err := r.Run(":8080"); err != nil {
			log.Printf("HTTP-server error: %v\n", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Println("Service is shutting down....")
	cancel()
	time.Sleep(5 * time.Second)
	log.Println("Service is down")

}
