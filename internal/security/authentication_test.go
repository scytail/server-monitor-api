package security

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"testing"
)

func TestReadAPIToken(t *testing.T) {
	directory := t.TempDir()
	want := base64.RawURLEncoding.EncodeToString(make([]byte, 32))
	if err := os.WriteFile(filepath.Join(directory, apiTokenCredentialName), []byte(want+"\n"), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CREDENTIALS_DIRECTORY", directory)

	got, err := ReadAPIToken()
	if err != nil {
		t.Fatalf("readAPIToken() error = %v, want nil", err)
	}
	if got != want {
		t.Fatalf("readAPIToken() = %q, want %q", got, want)
	}
}

func TestReadAPITokenRejectsMissingOrInvalidCredentials(t *testing.T) {
	t.Setenv("CREDENTIALS_DIRECTORY", "")
	if _, err := ReadAPIToken(); err == nil {
		t.Fatal("readAPIToken() error = nil for missing credentials directory")
	}

	tests := []struct {
		name string
		data string
	}{
		{name: "missing credential file"},
		{name: "invalid encoding", data: "not-a-token"},
		{name: "wrong token length", data: base64.RawURLEncoding.EncodeToString([]byte("short"))},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			directory := t.TempDir()
			t.Setenv("CREDENTIALS_DIRECTORY", directory)
			if tt.data != "" {
				if err := os.WriteFile(filepath.Join(directory, apiTokenCredentialName), []byte(tt.data), 0600); err != nil {
					t.Fatal(err)
				}
			}
			if _, err := ReadAPIToken(); err == nil {
				t.Fatal("readAPIToken() error = nil, want error")
			}
		})
	}
}
