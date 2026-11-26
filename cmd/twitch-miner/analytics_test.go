package main

import "testing"

func TestAnalyticsAddr(t *testing.T) {
	tests := []struct {
		name     string
		cfg      analyticsConfig
		override string
		want     string
	}{
		{name: "a host and a port", cfg: analyticsConfig{Host: "0.0.0.0", Port: 8080}, want: "0.0.0.0:8080"},
		{name: "nothing set", cfg: analyticsConfig{}, want: "127.0.0.1:5000"},
		{name: "no host", cfg: analyticsConfig{Port: 8080}, want: "127.0.0.1:8080"},
		{name: "a blank host", cfg: analyticsConfig{Host: "  ", Port: 8080}, want: "127.0.0.1:8080"},
		{name: "a padded host", cfg: analyticsConfig{Host: " localhost ", Port: 8080}, want: "localhost:8080"},
		{name: "an IPv6 host", cfg: analyticsConfig{Host: "::1", Port: 8080}, want: "[::1]:8080"},
		{name: "port zero", cfg: analyticsConfig{Host: "0.0.0.0", Port: 0}, want: "0.0.0.0:5000"},
		{name: "port one", cfg: analyticsConfig{Host: "0.0.0.0", Port: 1}, want: "0.0.0.0:1"},
		{name: "a negative port", cfg: analyticsConfig{Host: "0.0.0.0", Port: -1}, want: "0.0.0.0:5000"},
		{name: "an override", cfg: analyticsConfig{Host: "0.0.0.0", Port: 8080}, override: "127.0.0.1:9000", want: "127.0.0.1:9000"},
		{name: "a padded override", cfg: analyticsConfig{Host: "0.0.0.0", Port: 8080}, override: " :9000 ", want: ":9000"},
		{name: "a blank override", cfg: analyticsConfig{Host: "0.0.0.0", Port: 8080}, override: "   ", want: "0.0.0.0:8080"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := analyticsAddr(tt.cfg, tt.override); got != tt.want {
				t.Errorf("analyticsAddr(%+v, %q) = %q, want %q", tt.cfg, tt.override, got, tt.want)
			}
		})
	}
}
