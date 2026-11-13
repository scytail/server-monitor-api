package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"server-monitor-api/internal/services"
)

func TestServicesEndpoints(t *testing.T) {
	allServices := []services.Service{
		{Name: "alpha.service", Active: true},
		{Name: "zeta.service", Active: false},
	}
	handler := NewRouter(func(context.Context) ([]services.Service, error) {
		return allServices, nil
	})

	tests := []struct {
		name       string
		method     string
		body       string
		wantStatus int
		want       []services.Service
	}{
		{
			name:       "list services",
			method:     http.MethodGet,
			wantStatus: http.StatusOK,
			want:       allServices,
		},
		{
			name:       "filter services by names",
			method:     http.MethodPost,
			body:       `["zeta.service"]`,
			wantStatus: http.StatusOK,
			want:       []services.Service{{Name: "zeta.service", Active: false}},
		},
		{
			name:       "empty filter returns no services",
			method:     http.MethodPost,
			body:       `[]`,
			wantStatus: http.StatusOK,
			want:       []services.Service{},
		},
		{
			name:       "reject malformed JSON",
			method:     http.MethodPost,
			body:       `{`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "reject non-array JSON",
			method:     http.MethodPost,
			body:       `{"name":"alpha.service"}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "reject trailing JSON values",
			method:     http.MethodPost,
			body:       `[] []`,
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(tt.method, "/services", strings.NewReader(tt.body))
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)

			if response.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d; body: %s", response.Code, tt.wantStatus, response.Body)
			}
			if response.Header().Get("Content-Type") != "application/json" {
				t.Errorf("Content-Type = %q, want application/json", response.Header().Get("Content-Type"))
			}
			if tt.wantStatus != http.StatusOK {
				return
			}

			var got []services.Service
			if err := json.Unmarshal(response.Body.Bytes(), &got); err != nil {
				t.Fatalf("decode response JSON: %v", err)
			}
			if len(got) != len(tt.want) {
				t.Fatalf("services = %#v, want %#v", got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("services[%d] = %#v, want %#v", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestServicesEndpointReportsDiscoveryErrors(t *testing.T) {
	handler := NewRouter(func(context.Context) ([]services.Service, error) {
		return nil, errors.New("systemd unavailable")
	})

	for _, method := range []string{http.MethodGet, http.MethodPost} {
		request := httptest.NewRequest(method, "/services", strings.NewReader(`[]`))
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)

		if response.Code != http.StatusInternalServerError {
			t.Errorf("%s status = %d, want %d", method, response.Code, http.StatusInternalServerError)
		}
	}
}
