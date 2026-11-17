package api

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

func TestDecodeNames(t *testing.T) {
	tests := []struct {
		name    string
		body    string
		want    []string
		wantErr bool
	}{
		{name: "valid array", body: `["alpha.service","zeta.service"]`, want: []string{"alpha.service", "zeta.service"}},
		{name: "empty array", body: `[]`, want: []string{}},
		{name: "malformed json", body: `{`, wantErr: true},
		{name: "non-array json", body: `{"name":"alpha.service"}`, wantErr: true},
		{name: "trailing values", body: `[] []`, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/services", strings.NewReader(tt.body))
			var got []string
			err := decodeNames(request, &got)

			if tt.wantErr {
				if err == nil {
					t.Fatal("decodeNames() error = nil, want error")
				}
				return
			}

			if err != nil {
				t.Fatalf("decodeNames() error = %v, want nil", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("decodeNames() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestRouterRejectsMalformedFilterPayloads(t *testing.T) {
	handler := NewRouter()

	for _, body := range []string{"{", `{"name":"alpha.service"}`, `[] []`} {
		request := httptest.NewRequest(http.MethodPost, "/services", strings.NewReader(body))
		response := httptest.NewRecorder()

		handler.ServeHTTP(response, request)

		if response.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d; body: %s", response.Code, http.StatusBadRequest, response.Body.String())
		}
		if response.Header().Get("Content-Type") != "application/json" {
			t.Fatalf("Content-Type = %q, want application/json", response.Header().Get("Content-Type"))
		}
	}
}

func TestRouterLogsRequests(t *testing.T) {
	var output bytes.Buffer
	previousOutput := log.Writer()
	log.SetOutput(&output)
	t.Cleanup(func() {
		log.SetOutput(previousOutput)
	})

	request := httptest.NewRequest(http.MethodPost, "/services", strings.NewReader("{"))
	response := httptest.NewRecorder()

	NewRouter().ServeHTTP(response, request)

	if !strings.Contains(output.String(), "API request: POST /services") {
		t.Fatalf("log output = %q, want request method and path", output.String())
	}
}
