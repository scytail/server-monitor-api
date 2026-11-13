package main

import (
	"flag"
	"log"
	"net"
	"net/http"
	"strconv"

	"server-monitor-api/internal/api"
	"server-monitor-api/internal/services"
)

func main() {
	ip := flag.String("ip", "localhost", "IP address or host name to listen on")
	port := flag.Int("port", 8765, "port to listen on")
	flag.Parse()

	address := net.JoinHostPort(*ip, strconv.Itoa(*port))
	log.Printf("API listening on %s", address)
	log.Fatal(http.ListenAndServe(address, api.NewRouter(services.ListServices)))
}
