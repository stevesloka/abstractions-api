/*
	Migrations: https://github.com/mattes/migrate
	MySQL: https://github.com/go-sql-driver/mysql
*/

package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/stevesloka/abstractions-api/healthz"

	"database/sql"

	gorilla "github.com/gorilla/http"
	"github.com/gorilla/mux"
	_ "github.com/mattes/migrate/driver/mysql"
	"github.com/mattes/migrate/migrate"
)

var (
	argListenPort    = flag.Int("listen-port", 80, "port to have API listen")
	argTLSListenPort = flag.Int("tls-listen-port", 9443, "port to have API listen over TLS")
	argDBUserName    = os.Getenv("DATABASE_USERNAME")
	argDBPassword    = os.Getenv("DATABASE_PASSWORD")
	argDBPort        = os.Getenv("DATABASE_PORT")
	argDBHost        = os.Getenv("DATABASE_HOST")
	argDBName        = os.Getenv("DATABASE_NAME")
	db               *sql.DB
)

const (
	appVersion = "0.0.1"
)

type test_struct struct {
	Test string
}

// Organizers (GET "/organizers")
func organizers(w http.ResponseWriter, r *http.Request) {
	result, _ := getJSON("SELECT * FROM organizer")
	fmt.Fprintf(w, result)
}

// Organizers (GET "/sessions")
func sessions(w http.ResponseWriter, r *http.Request) {
	buf := new(bytes.Buffer)
	if _, err := gorilla.Get(buf, "http://abstractions.io/api/schedule.json"); err != nil {
		log.Fatalf("could not fetch: %v", err)
	}

	fmt.Fprintf(w, buf.String())
}

// Version (GET "/version")
func versionRoute(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "%q", appVersion)
}

func getJSON(sqlString string) (string, error) {

	// Configure MySQL
	connectionDSN := fmt.Sprintf("%s:%s@(%s:%s)/%s", argDBUserName, argDBPassword, argDBHost, argDBPort, argDBName)

	db, err := sql.Open("mysql", connectionDSN)
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

func validateEnvVariables() {
	if argDBUserName == "" {
		argDBUserName = "root"
	}

	if argDBPassword == "" {
		argDBPassword = "password"
	}

	if argDBPort == "" {
		argDBPort = "3306"
	}

	if argDBHost == "" {
		argDBHost = "127.0.0.1"
	}

	if argDBName == "" {
		argDBName = "abstractions"
	}
}

func main() {
	flag.Parse()
	log.Println("API is up and running!", time.Now(), " running on port: ", *argListenPort)

	// Validated environment vars and set defaults if needed
	validateEnvVariables()

	// Get hostname
	hostname, err := os.Hostname()
	if err != nil {
		log.Fatal(err)
	}

	// Configure MySQL
	connectionDSN := fmt.Sprintf("%s:%s@(%s:%s)/%s", argDBUserName, argDBPassword, argDBHost, argDBPort, argDBName)

	hc := &healthz.Config{
		Hostname: hostname,
		Database: healthz.DatabaseConfig{
			DriverName:     "mysql",
			DataSourceName: connectionDSN,
		},
	}

	healthzHandler, err := healthz.Handler(hc)
	if err != nil {
		log.Fatal(err)
	}

	// Configure router
	router := mux.NewRouter().StrictSlash(true)

	// Routes
	router.HandleFunc("/version", versionRoute)
	router.HandleFunc("/organizers", organizers)
	router.HandleFunc("/sessions", sessions)
	router.Handle("/healthz", healthzHandler)

	// Start insecure server
	go func() {
		err_http := http.ListenAndServe(fmt.Sprintf(":%d", *argListenPort), router)
		if err_http != nil {
			log.Fatal("Web server (HTTP): ", err_http)
		}
	}()

	canConnect := true

	// Configure MySQL
	db, err := sql.Open("mysql", connectionDSN)
	if err != nil {
		canConnect = false
	}
	defer db.Close()

	// Open doesn't open a connection. Validate DSN data:
	err = db.Ping()
	if err != nil {
		canConnect = false
	}

	for canConnect == false {
		time.Sleep(time.Second * 3)

		// Configure MySQL
		db, err := sql.Open("mysql", connectionDSN)
		defer db.Close()

		// Open doesn't open a connection. Validate DSN data:
		err = db.Ping()
		if err == nil {
			canConnect = true
		}
	}

	// Run Migrations
	// use synchronous versions of migration functions ...
	log.Println("About to run migrations....")
	errors, ok := migrate.UpSync(fmt.Sprintf("mysql://%s", connectionDSN), "./db/migrations")
	if !ok {
		log.Println("Oh no ... migrations failed!")
		// do sth with allErrors slice
		for err := range errors {
			log.Println(err)
		}

		log.Println(errors)
		panic("Bailing out on running migrations!")
	}

	log.Println("Migrations run!")

	// Start secure server
	log.Fatal(http.ListenAndServeTLS(fmt.Sprintf(":%d", *argTLSListenPort), "certs/cert.pem", "certs/key.pem", router))
}
