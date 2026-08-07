package config

import "testing"

func TestAddressFromURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		rawURL      string
		defaultPort string
		want        string
	}{
		{name: "postgres", rawURL: "postgres://user:pass@db:5433/app", defaultPort: "5432", want: "db:5433"},
		{name: "redis default port", rawURL: "redis://cache/0", defaultPort: "6379", want: "cache:6379"},
		{name: "ipv6", rawURL: "redis://[::1]:6380/0", defaultPort: "6379", want: "[::1]:6380"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := addressFromURL(tt.rawURL, tt.defaultPort)
			if err != nil {
				t.Fatalf("addressFromURL() error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("addressFromURL() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestAddressFromURLRejectsMissingHost(t *testing.T) {
	t.Parallel()
	if _, err := addressFromURL("postgres:///codebattle", "5432"); err == nil {
		t.Fatal("addressFromURL() expected an error")
	}
}
