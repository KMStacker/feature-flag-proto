package main

import (
	"database/sql"  // for SQL database interaction
	"encoding/json" // for parsing JSON
	"fmt"           // for console prints
	"log"           // for logging errors
	"net/http"      // for creating the HTTP server
	"os"            // for env vars
	"sync"          // for syncing mutexes
	"time"          // for time hacks

	_ "github.com/lib/pq" // PostgreSQL translator for registering the driver
)

// global variables
var (
	featureFlag = false      // starting state of the feature flag in RAM
	mu          sync.RWMutex // mu(tex) lock for preventing race conditions when writing the flag (allowe multiple readers, single writer)
	db          *sql.DB      // memory address (=*) for the PostgreSQL database connection
)

// this defines the structure of the flag update requests from the flag-management frontend
type FlagUpdate struct {
	State bool `json:"state"`
}

// database connection function with retry loop
func connectDB() {
	connStr := os.Getenv("DB_CONN_STR") // getting the connection string from env variable

	// in case env variable is empty
	if connStr == "" {
		fmt.Println("No DB_CONN_STR found, using default local connection string.")           // log message
		connStr = "postgres://postgres:salasana@localhost:5432/feature_flags?sslmode=disable" // reserve connection string
	}

	var err error // creates err var for the loop below
	// loop for connection retries
	for i := 0; i < 10; i++ {

		db, err = sql.Open("postgres", connStr) // open connection, connection test incoming...

		// connection test
		if err == nil { // if there are no errors opening connection
			err = db.Ping() // pinging db to test connection
			if err == nil { // if successful ping -> all good n everything works
				fmt.Println("Connected to this masterpiece database with ease!") // success message for console
				return
			}
		}

		// if connection failed
		fmt.Println("Waiting for database connection...") // log message
		time.Sleep(2 * time.Second)                       // wait time before retrying
	}

	// if 1 try and 9 retries are not enough, then some fatal logging, error and sayonara!
	log.Fatal("Could not connect to database:", err)
}

// initFlagState ensures the database table exists and loads the initial state into RAM.
func initFlagState() {

	// creates the tble if it doesn't exist
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS flags (name VARCHAR(99) PRIMARY KEY, enabled BOOLEAN)`) // ("_" ignores the first value in return), (":=" creates a new var and gives it a value), table name: flags, columns: name (max 99 chars) & enabled (true/false)
	if err != nil {                                                                                       // if number of errors is not zero...
		log.Fatal("Failed to create table:", err) // error log and exit
	}

	// sets the default flag state if it doesn't exist yet
	_, err = db.Exec(`INSERT INTO flags (name, enabled) VALUES ('feature-flag-1', false) ON CONFLICT (name) DO NOTHING`) // (if var "err" already exists, use "=" instead of ":="), inserts default flag value (false) if there is no existing row with the same name
	if err != nil {                                                                                                      // if number of errors is not zero...
		log.Fatal("Initialization failed:", err) // error log and exit
	}

	// gets the current flag state from the database for the RAM
	var enabled bool                                                                          // creates a new var for the flag state
	err = db.QueryRow("SELECT enabled FROM flags WHERE name='feature-flag-1'").Scan(&enabled) // if all good, "Scan" puts the fetched value into the "enabled" var using memory address (=&) and err is nil, otherwise err got some value

	if err != nil { // if there was an error fetching the flag state
		log.Printf("Warning: Could not fetch flag state from DB: %v", err) // continues execution with default flag value but logs a warning
	} else { // uses locks and updates featureFlag var value (for the RAM) if everything went well
		mu.Lock()
		featureFlag = enabled
		mu.Unlock()
		fmt.Printf("This such a nice journey just started. Flag state loaded from DB: %v\n", featureFlag) // log message for the sake of information about the initial flag state in the beginning
	}
}

// middleware to enable cross-origin requests
func enableCORS(next http.Handler) http.Handler { // next is the next handler in the chain, returns a modified handler
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { // chages the handler function with inside function logic
		w.Header().Set("Access-Control-Allow-Origin", "*")                   // allow requests from any origin
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS") // allow said methods
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")       // allow content-type header for JSON type things

		// "preflight request" for rules above
		if r.Method == "OPTIONS" {
			return
		}

		next.ServeHTTP(w, r) // pass the request to the next handler in the chain
	})
}

// flagHandler handles requests to /api/flags
func flagHandler(w http.ResponseWriter, r *http.Request) { // w is the response writer, r is the incoming request
	w.Header().Set("Content-Type", "application/json") // sets the response content type to JSON

	// POST requests
	if r.Method == http.MethodPost { // if the request method is POST...
		var update FlagUpdate // creates an empty struct for incoming data
		// decodes the JSON request body into the struct if err is nil, otherwise returns bad request error
		if err := json.NewDecoder(r.Body).Decode(&update); err != nil { // uses ";" to create new var "err" for only inside this "if" block
			http.Error(w, "Invalid request data", http.StatusBadRequest)
			return
		}

		// if not returned, locks for writing and updates RAM
		mu.Lock()
		featureFlag = update.State
		mu.Unlock()

		// updates the database
		_, err := db.Exec("UPDATE flags SET enabled=$1 WHERE name='feature-flag-1'", update.State) // for safety reasons we use "$1" as a placeholder for the first argument after the query string
		if err != nil {                                                                            // if there was an error updating the database...
			log.Printf("ERROR: Database update failed: %v", err)
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}

		// if not returned earlier, log success and return good news to frontend
		fmt.Printf("Flag state saved to DB and memory: %v\n", featureFlag)
		json.NewEncoder(w).Encode(map[string]bool{"success": true}) // good news = {"success": true}
		return                                                      // exit before reaching the read operation GET part
	}

	// GET requests
	mu.RLock()                  // read lock for readers which allows multiple readers simultaneously
	currentState := featureFlag // create a new var by reading directly from RAM (no interaction with DB)
	mu.RUnlock()                // unlocks the read lock

	json.NewEncoder(w).Encode(map[string]bool{"feature-flag-1": currentState}) // encodes the current flag state to JSON and sends it as the response
}

func main() {
	connectDB()     // establishes the connection to PostgreSQL
	initFlagState() // loads the last known state from DB to RAM

	mux := http.NewServeMux() // creates a new router (multiplexer=mux)

	mux.HandleFunc("/api/flags", flagHandler) // register the endpoint and its handler

	// checks the health of the endpoints
	mux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	fmt.Println("Server running in port 8080...") // log message for server start

	http.ListenAndServe(":8080", enableCORS(mux)) // starts the HTTP server with CORS middleware enabled and including the mux router
}
