package security

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const apiTokenCredentialName = "gomonitor-api-token"

func ReadAPIToken() (string, error) {
	credentialsDirectory := os.Getenv("CREDENTIALS_DIRECTORY")
	if credentialsDirectory == "" {
		return "", errors.New("CREDENTIALS_DIRECTORY is not set")
	}

	path := filepath.Join(credentialsDirectory, apiTokenCredentialName)
	credential, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read API token credential %q: %w", apiTokenCredentialName, err)
	}

	token := strings.TrimRight(string(credential), "\r\n")
	decoded, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil || len(decoded) != 32 {
		return "", fmt.Errorf("API token credential %q must contain a 32-byte URL-safe base64 token", apiTokenCredentialName)
	}
	return token, nil
}

func RequireBearerToken(token string, next http.Handler) http.Handler {
	expected := sha256.Sum256([]byte(token))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		headers := r.Header.Values("Authorization")
		if len(headers) != 1 {
			writeUnauthorized(w)
			return
		}

		parts := strings.Fields(headers[0])
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			writeUnauthorized(w)
			return
		}

		provided := sha256.Sum256([]byte(parts[1]))
		if subtle.ConstantTimeCompare(expected[:], provided[:]) != 1 {
			writeUnauthorized(w)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func writeUnauthorized(w http.ResponseWriter) {
	w.Header().Set("WWW-Authenticate", "Bearer")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	if err := json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"}); err != nil {
		log.Printf("encode API response: %v", err)
	}
}
