package main

import (
	"flag"
	"log"
	"net"
	"net/http"
	"strconv"

	"server-monitor-api/api"
	"server-monitor-api/internal/security"
)

func main() {
	ip := flag.String("ip", "localhost", "IP address or host name to listen on")
	port := flag.Int("port", 8765, "port to listen on")
	flag.Parse()

	token, err := security.ReadAPIToken()
	if err != nil {
		log.Fatal(err)
	}

	address := net.JoinHostPort(*ip, strconv.Itoa(*port))
	log.Printf("API listening on %s", address)
	log.Fatal(http.ListenAndServe(address, api.NewRouter(token)))
}
