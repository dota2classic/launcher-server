package main

import (
	"log"
	"os"

	"awesomeProject/internal/server"
)

func main() {
	basePath := os.Getenv("LAUNCHER_FILES_PATH")
	if basePath == "" {
		log.Fatal("LAUNCHER_FILES_PATH environment variable is not set")
	}

	addr := os.Getenv("LAUNCHER_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	srv, err := server.New(basePath)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	log.Fatal(srv.ListenAndServe(addr))
}
