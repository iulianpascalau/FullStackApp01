package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"FullStackApp01/api"
	"FullStackApp01/storage"

	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	jwtKey := os.Getenv("JWT_KEY")
	adminPassword := os.Getenv("ADMIN_PASSWORD")

	if jwtKey == "" || adminPassword == "" {
		log.Fatal("JWT_KEY and ADMIN_PASSWORD must be set in .env")
	}

	// Create or open a database in the "data" folder
	store, err := storage.NewStore("data")
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer func() {
		_ = store.Close()
	}()

	// Ensure an admin exists
	_ = store.SaveUser("admin", adminPassword, "admin")

	server := api.NewServer(store, []byte(jwtKey))

	http.HandleFunc("/register", server.HandleRegister)
	http.HandleFunc("/login", server.HandleLogin)
	http.HandleFunc("/change-password", server.HandleChangePassword)
	http.HandleFunc("/counter", server.HandleCounter)

	port := ":8080"
	fmt.Printf("Starting server on port %s\n", port)
	err = http.ListenAndServe(port, nil)
	if err != nil {
		log.Fatalf("Could not start server: %s\n", err)
	}
}
