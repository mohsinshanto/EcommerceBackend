package main

import (
	"context"
	"ecommerce-backend/config"
	"ecommerce-backend/services"
	"flag"
	"log"
)

func main() {
	batchSize := flag.Int("batch", 500, "number of products to sync per batch")
	flag.Parse()

	config.LoadEnv()
	config.ConnectDB()
	config.InitTypesense()

	if config.Typesense == nil {
		log.Fatal("Typesense is not enabled. Set TYPESENSE_API_KEY and retry.")
	}

	log.Println("Syncing MySQL products to Typesense...")
	if err := services.SyncAllProductsToTypesense(context.Background(), *batchSize); err != nil {
		log.Fatalf("failed to sync products to Typesense: %v", err)
	}

	log.Println("Typesense product sync complete.")
}
