package main

import (
	"log"
	"net/http"
	"os"
	"strconv"

	httphandler "github.com/aleksandr/strive-api/internal/http"
)

func getEnvOrDefault(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func parsePort(portStr string) (int, error) {
	return strconv.Atoi(portStr)
}

func main() {
	portStr := getEnvOrDefault("PORT", "8080")
	port, err := parsePort(portStr)
	if err != nil || port <= 0 || port > 65535 {
		log.Fatalf("invalid PORT: %s", portStr)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", httphandler.HealthHandler)

	server := httphandler.NewServer(port, mux)
	server.Start()
	server.WaitForShutdown()
}
