package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/golang/glog"
	"github.com/gorilla/mux"
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
	glog.V(4).Info("API is up and running!", time.Now())

	// Configure router
	router := mux.NewRouter().StrictSlash(true)

	// Version
	router.HandleFunc("/version", versionRoute)

	// Start server
	log.Fatal(http.ListenAndServeTLS(fmt.Sprintf(":%d", *argListenPort), "certs/cert.pem", "certs/key.pem", router))
}
