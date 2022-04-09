package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

type Config struct {
	port uint
}

func parseFlags() Config {
	port := flag.Uint("port", 8080, "server port")
	return Config{
		*port,
	}
}

func main() {
	config := parseFlags()

	r := mux.NewRouter()
	r.HandleFunc("/r/{room}", roomHandler)
	http.Handle("/", r)

	log.Println("server listening on port", config.port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf("localhost:%d", config.port), nil))
}
