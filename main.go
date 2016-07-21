/*
	Migrations: https://github.com/mattes/migrate
	MySQL: https://github.com/go-sql-driver/mysql
*/

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"database/sql"

	"github.com/gorilla/mux"
	_ "github.com/mattes/migrate/driver/mysql"
	"github.com/mattes/migrate/migrate"
)

var (
	argListenPort = flag.Int("listen-port", 9080, "port to have API listen")
	db            *sql.DB
)

const (
	appVersion = "0.0.1"
)

// Organizers (GET "/organizers")
func organizers(w http.ResponseWriter, r *http.Request) {
	result, _ := getJSON("SELECT * FROM organizer")
	fmt.Fprintf(w, result)
}

// Version (GET "/version")
func versionRoute(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "%q", appVersion)
}

func getJSON(sqlString string) (string, error) {
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

	stmt, err := db.Prepare(sqlString)
	if err != nil {
		return "", err
	}
	defer stmt.Close()

	rows, err := stmt.Query()
	if err != nil {
		return "", err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return "", err
	}

	tableData := make([]map[string]interface{}, 0)

	count := len(columns)
	values := make([]interface{}, count)
	scanArgs := make([]interface{}, count)
	for i := range values {
		scanArgs[i] = &values[i]
	}

	for rows.Next() {
		err := rows.Scan(scanArgs...)
		if err != nil {
			return "", err
		}

		entry := make(map[string]interface{})
		for i, col := range columns {
			v := values[i]

			b, ok := v.([]byte)
			if ok {
				entry[col] = string(b)
			} else {
				entry[col] = v
			}
		}

		tableData = append(tableData, entry)
	}

	jsonData, err := json.Marshal(tableData)
	if err != nil {
		return "", err
	}

	return string(jsonData), nil
}

func main() {
	flag.Parse()
	fmt.Println("API is up and running!", time.Now())

	// Configure router
	router := mux.NewRouter().StrictSlash(true)

	// Version
	router.HandleFunc("/version", versionRoute)
	router.HandleFunc("/organizers", organizers)

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
