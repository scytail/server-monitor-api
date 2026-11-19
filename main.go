package main

import (
	"flag"
	"log"
	"net"
	"net/http"
	"strconv"

	"server-monitor-api/api"
	"server-monitor-api/internal/security"
	"server-monitor-api/internal/services"
)

func main() {
	ip := flag.String("ip", "localhost", "IP address or host name to listen on")
	port := flag.Int("port", 8765, "port to listen on")
	blacklistFile := flag.String("blacklist-file", "/etc/server-monitor-api/blacklist", "path to the service blacklist file")
	flag.Parse()

	token, err := security.ReadAPIToken()
	if err != nil {
		log.Fatal(err)
	}

	blacklist, found, err := services.LoadBlacklist(*blacklistFile)
	if err != nil {
		log.Fatal(err)
	}
	if found {
		log.Printf("loaded service blacklist from %s", *blacklistFile)
	} else {
		log.Printf("service blacklist file %s not found; using an empty blacklist", *blacklistFile)
	}

	address := net.JoinHostPort(*ip, strconv.Itoa(*port))
	log.Printf("API listening on %s", address)
	log.Fatal(http.ListenAndServe(address, api.NewRouter(token, blacklist...)))
}
