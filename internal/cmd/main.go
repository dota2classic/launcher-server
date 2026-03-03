package main

import (
	"log"
	"os"

	"launcher-host/internal/server"
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

	certFile := os.Getenv("LAUNCHER_TLS_CERT")
	keyFile := os.Getenv("LAUNCHER_TLS_KEY")

	srv, err := server.New(basePath)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	if certFile != "" && keyFile != "" {
		log.Fatal(srv.ListenAndServeTLS(addr, certFile, keyFile))
	} else {
		log.Fatal(srv.ListenAndServe(addr))
	}
}
