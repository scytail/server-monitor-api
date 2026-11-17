package api

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net"
	"net/http"

	"server-monitor-api/internal/services"
)

type ServiceController func(context.Context) ([]services.Service, error)

func NewRouter() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /services", listServicesHandler)
	mux.HandleFunc("POST /services", filterServicesHandler)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientIP := r.RemoteAddr
		if host, _, err := net.SplitHostPort(clientIP); err == nil {
			clientIP = host
		}
		if ip := net.ParseIP(clientIP); ip != nil {
			log.Printf("API request from %s: %s %s", ip, r.Method, r.URL.Path)
		} else {
			log.Printf("API request: %s %s", r.Method, r.URL.Path)
		}
		mux.ServeHTTP(w, r)
	})
}

func listServicesHandler(w http.ResponseWriter, r *http.Request) {
	allServices, err := services.GetAllServices(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list services")
		log.Printf("list services: %v", err)
		return
	}

	writeJSON(w, http.StatusOK, allServices)
}

func filterServicesHandler(w http.ResponseWriter, r *http.Request) {
	var names []string
	if err := decodeNames(r, &names); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	allServices, err := services.GetAllServices(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list services")
		log.Printf("list services: %v", err)
		return
	}

	writeJSON(w, http.StatusOK, services.FilterServices(allServices, names))
}

func decodeNames(r *http.Request, names *[]string) error {
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(names); err != nil {
		return errors.New("request body must be a JSON array of service names")
	}
	if *names == nil {
		return errors.New("request body must be a JSON array of service names")
	}

	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		return errors.New("request body must contain only one JSON value")
	}
	return nil
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		log.Printf("encode API response: %v", err)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
