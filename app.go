package main

import (
	"fmt"
	"log"
	"os"
	"database/sql"
	"html/template"
	"net/http"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3" // Import package just for its declarations
)

const DEBUG = true

func initDB(dbPath string) *sql.DB {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatal(err)
	}

	// Create table ONLY if it does not exist
	const createTableQuery = `
	CREATE TABLE IF NOT EXISTS confirm(
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		session_id VARCHAR(255) NOT NULL UNIQUE,
		is_confirmed BOOLEAN DEFAULT 0
	);
	`
	
	_, err = db.Exec(createTableQuery)
	if err != nil {
		log.Fatal(err)
	}

	return db
}

func loadTemplate(name string) *template.Template {
	/* Loads a template from disk and returns it */
	tmpl, err := template.ParseFiles(filepath.Join("templates", name + ".html"))
	if err != nil {
		log.Fatalf("Template failed to execute: %s", err)
	}
	
	return tmpl
}

func handleIndex(db *sql.DB) http.HandlerFunc {
	/* Returns a function that serves the index page */
	return func(w http.ResponseWriter, r *http.Request) {
		// Load index template from disk and parse it
		tmpl := loadTemplate("index")
		err := tmpl.ExecuteTemplate(w, "index", nil)
		if err != nil {
			log.Fatalf("Template failed to execute: %s", err)
		}
	}
}

func handleViewConfirmation(db *sql.DB) http.HandlerFunc {
	/* Returns a function that serves the confirm page */
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()

		var sessionID string
		success := false

		// Retrieve session ID from arguments
		if val, ok := q["s"]; ok {
			// Look up ID in database
			rows, err := db.Query("SELECT session_id FROM confirm WHERE session_id=?", val[0])
			if err != nil {
				log.Fatal("DB lookup for session ID failed!")
			}

			rows.Scan(&sessionID)

			success = true
		}

		if success {
			// Display confirmation page
			tmpl := loadTemplate("confirm")

			// Pass sessionID into template (use args struct for easier addition of arguments)
			args := struct {
				SessionID string
			}{sessionID}
			
			err := tmpl.ExecuteTemplate(w, "confirm", args)
			if err != nil {
				log.Fatal("Error parsing confirm template!")
			}
		} else {
			// Session ID invalid
			tmpl := loadTemplate("404")
			tmpl.ExecuteTemplate(w, "404", nil)
		}
	}
}

func handleConfirm(db *sql.DB) http.HandlerFunc {
	/* Handles an incoming confirmation and displays a success page */
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()

		var sessionID string
		success := false

		// Retrieve session ID from arguments
		if val, ok := q["s"]; ok {
			// Look up ID in database
			rows, err := db.Query("SELECT session_id FROM confirm WHERE session_id=?", val[0])
			if err != nil {
				log.Fatal("DB lookup for session ID failed!")
			}
			
			rows.Scan(&sessionID)

			success = true
		}

		if success {
			// Load confirm template from disk and parse it
		tmpl := loadTemplate("confirm")
		err := tmpl.ExecuteTemplate(w, "confirm", nil)
		if err != nil {
			log.Fatalf("Template failed to execute: %s", err)
		}
		}
		

		
	}
}

func startServer(db *sql.DB) {
	http.HandleFunc("/", handleIndex(db))
	http.HandleFunc("/view", handleViewConfirmation(db))
	http.HandleFunc("/confirm", handleConfirm(db))
	log.Fatal(http.ListenAndServe(":8888", nil))
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: confirm <path_to_db>")
		os.Exit(0)
	}

	// Initialize database
	dbPath := os.Args[1]
	db := initDB(dbPath)

	log.Println("DB loaded successfully")
	log.Println("Server listening on 8080...")

	// Start the actual HTTP server, running locally
	startServer(db)
}
