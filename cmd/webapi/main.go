package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/mattn/go-sqlite3"
	"github.com/rerikdev/wasatext/service/api"
	"github.com/rerikdev/wasatext/service/database"
)

func main() {
	// Open database
	sqliteDb, err := sql.Open("sqlite3", "./wasatext.db")
	if err != nil {
		log.Printf("Error opening database: %v", err)
		os.Exit(1)
	}
	defer sqliteDb.Close()

	// Initialize database logic
	db, err := database.New(sqliteDb)
	if err != nil {
		log.Printf("Error initializing database logic: %v", err)
		os.Exit(1)
	}

	// Initialize API router
	router, err := api.New(db)
	if err != nil {
		log.Printf("Error creating API instance: %v", err)
		os.Exit(1)
	}

	// ---------- Serve Vue static files ----------
	// Serve everything under /public/ (the built frontend)
	router.Router.ServeFiles("/public/*filepath", http.Dir("./public"))

	// Optional: Redirect root (/) to /public/index.html for SPA routing
	router.Router.NotFound = http.FileServer(http.Dir("./public"))
	// ---------------------------------------------

	port := ":3000"
	fmt.Println("WASAText Backend is running on http://localhost" + port)

	// Start server
	err = http.ListenAndServe(port, router.Handler())
	if err != nil {
		log.Printf("Error starting server: %v", err)
		os.Exit(1)
	}
}