package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"database/sql"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	_ "github.com/mattes/migrate/driver/mysql"
	"github.com/mattes/migrate/migrate"
)

var (
	argListenPort = flag.Int("listen-port", 9080, "port to have API listen")
)

const (
	appVersion = "0.0.1"
)

// Version (GET "/version")
func versionRoute(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "%q", appVersion)
}

func main() {
	flag.Parse()
	fmt.Println("API is up and running!", time.Now())

	// Configure router
	router := mux.NewRouter().StrictSlash(true)

	// Version
	router.HandleFunc("/version", versionRoute)

	// Configure MySQL
	db, err := sql.Open("mysql", "root:password@(127.0.0.1:3306)/abstractions")
	if err != nil {
		panic(err.Error()) // Just for example purpose. You should use proper error handling instead of panic
	}
	defer db.Close()

	// Open doesn't open a connection. Validate DSN data:
	err = db.Ping()
	if err != nil {
		panic(err.Error()) // proper error handling instead of panic in your app
	}

	// Run Migrations
	// use synchronous versions of migration functions ...
	errors, ok := migrate.UpSync("mysql://root:password@(127.0.0.1:3306)/abstractions", "./db/migrations")
	if !ok {
		fmt.Println("Oh no ... migrations failed!")
		// do sth with allErrors slice
		for err := range errors {
			fmt.Println(err)
		}

		fmt.Println(errors)
		panic("Bailing out on running migrations!")
	}

	// Start server
	log.Fatal(http.ListenAndServeTLS(fmt.Sprintf(":%d", *argListenPort), "certs/cert.pem", "certs/key.pem", router))
}
