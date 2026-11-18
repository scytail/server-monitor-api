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
	handler := NewRouter(testToken)

	for _, body := range []string{"{", `{"name":"alpha.service"}`, `[] []`} {
		request := httptest.NewRequest(http.MethodPost, "/services", strings.NewReader(body))
		request.Header.Set("Authorization", "Bearer "+testToken)
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

func TestRouterRequiresBearerToken(t *testing.T) {
	tests := []struct {
		name        string
		credentials []string
	}{
		{name: "missing token"},
		{name: "wrong token", credentials: []string{"Bearer invalid"}},
		{name: "wrong scheme", credentials: []string{"Basic " + testToken}},
		{name: "duplicate authorization headers", credentials: []string{"Bearer " + testToken, "Bearer " + testToken}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, endpoint := range []struct {
				method string
				path   string
			}{
				{method: http.MethodGet, path: "/services"},
				{method: http.MethodPost, path: "/services"},
			} {
				request := httptest.NewRequest(endpoint.method, endpoint.path, nil)
				for _, credential := range tt.credentials {
					request.Header.Add("Authorization", credential)
				}
				response := httptest.NewRecorder()

				NewRouter(testToken).ServeHTTP(response, request)

				if response.Code != http.StatusUnauthorized {
					t.Errorf("%s status = %d, want %d", endpoint.method, response.Code, http.StatusUnauthorized)
				}
				if response.Header().Get("WWW-Authenticate") != "Bearer" {
					t.Errorf("%s WWW-Authenticate = %q, want Bearer", endpoint.method, response.Header().Get("WWW-Authenticate"))
				}
			}
		})
	}
}

func TestRouterAcceptsBearerToken(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/services", strings.NewReader("{"))
	request.Header.Set("Authorization", "Bearer "+testToken)
	response := httptest.NewRecorder()

	NewRouter(testToken).ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
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
	request.RemoteAddr = "203.0.113.8:54321"
	request.Header.Set("Authorization", "Bearer "+testToken)
	response := httptest.NewRecorder()

	NewRouter(testToken).ServeHTTP(response, request)

	if !strings.Contains(output.String(), "API request from 203.0.113.8: POST /services") {
		t.Fatalf("log output = %q, want request IP, method, and path", output.String())
	}
	if strings.Contains(output.String(), testToken) {
		t.Fatal("request log contains the bearer token")
	}
}

const testToken = "test-token"
