package tox

import (
	"os"
	"testing"
)

func TestExtractHost(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"abc123.onion:1234", "abc123.onion"},
		{"xyz789.b32.i2p:5678", "xyz789.b32.i2p"},
		{"localhost:8080", "localhost"},
		{"example.onion", "example.onion"}, // no port
		{"plain.b32.i2p", "plain.b32.i2p"}, // no port
		{"192.168.1.1:443", "192.168.1.1"},
		{"[::1]:8080", "::1"}, // IPv6 with port
	}

	for _, tt := range tests {
		got := extractHost(tt.input)
		if got != tt.want {
			t.Errorf("extractHost(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestIsAnonOnlyMode(t *testing.T) {
	// Save original env value.
	original, hadOriginal := os.LookupEnv("MTOX_ANON_ONLY")
	defer func() {
		if hadOriginal {
			os.Setenv("MTOX_ANON_ONLY", original)
		} else {
			os.Unsetenv("MTOX_ANON_ONLY")
		}
	}()

	os.Unsetenv("MTOX_ANON_ONLY")
	if IsAnonOnlyMode() {
		t.Error("IsAnonOnlyMode() = true when MTOX_ANON_ONLY is unset")
	}

	os.Setenv("MTOX_ANON_ONLY", "0")
	if IsAnonOnlyMode() {
		t.Error("IsAnonOnlyMode() = true when MTOX_ANON_ONLY=0")
	}

	os.Setenv("MTOX_ANON_ONLY", "1")
	if !IsAnonOnlyMode() {
		t.Error("IsAnonOnlyMode() = false when MTOX_ANON_ONLY=1")
	}
}

func TestProfilePath(t *testing.T) {
	path := ProfilePath()
	if path == "" {
		t.Error("ProfilePath() returned empty string")
	}
	// Path should end with profile.tox
	if len(path) < 11 || path[len(path)-11:] != "profile.tox" {
		t.Errorf("ProfilePath() = %q, expected to end with 'profile.tox'", path)
	}
}
