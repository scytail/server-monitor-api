package api

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"

	"server-monitor-api/internal/services"
)

type ServiceHandler func(context.Context) ([]services.Service, error)

func NewRouter(serviceController ServiceHandler) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /services", listServicesHandler(serviceController))
	mux.HandleFunc("POST /services", filterServicesHandler(serviceController))
	return mux
}

func listServicesHandler(serviceController ServiceHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		allServices, err := serviceController(r.Context())
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to list services")
			log.Printf("list services: %v", err)
			return
		}

		writeJSON(w, http.StatusOK, allServices)
	}
}

func filterServicesHandler(serviceController ServiceHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var names []string
		if err := decodeNames(r, &names); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		allServices, err := serviceController(r.Context())
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to list services")
			log.Printf("list services: %v", err)
			return
		}

		writeJSON(w, http.StatusOK, services.FilterServices(allServices, names))
	}
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
