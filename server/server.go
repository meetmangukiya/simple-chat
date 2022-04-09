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
	host string
}

func parseFlags() Config {
	port := flag.Uint("port", 8080, "server port")
	host := flag.String("host", "0.0.0.0", "server host")
	flag.Parse()

	return Config{
		*port,
		*host,
	}
}

func main() {
	config := parseFlags()

	r := mux.NewRouter()
	r.HandleFunc("/r/{room}", roomHandler)
	http.Handle("/", r)

	log.Println("server listening on port", config.port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf("%s:%d", config.host, config.port), nil))
}
