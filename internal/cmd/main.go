package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

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

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGHUP)
	go func() {
		for range sigCh {
			log.Println("SIGHUP received, recalculating manifest...")
			if err := srv.Recalculate(); err != nil {
				log.Printf("Recalculate failed: %v", err)
			}
		}
	}()

	if certFile != "" && keyFile != "" {
		log.Fatal(srv.ListenAndServeTLS(addr, certFile, keyFile))
	} else {
		log.Fatal(srv.ListenAndServe(addr))
	}
}
