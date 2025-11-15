package main

import (
	"fmt"
	"log"
	"mor80/service-reviewer/internal/config"
)

func main() {
	fmt.Println("Starting service-reviewer...")

	cfg, err := config.Load("./configs/default.yaml")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	fmt.Printf("Configuration loaded: %+v\n", cfg)
}