package main

import "testing"

func TestDashboardAddrOverride(t *testing.T) {
	tests := []struct {
		name string
		env  map[string]string
		want string
	}{
		{name: "neither variable", env: map[string]string{}, want: ""},
		{name: "only TCPM_DASHBOARD_ADDR", env: map[string]string{"TCPM_DASHBOARD_ADDR": "0.0.0.0:8080"}, want: "0.0.0.0:8080"},
		{name: "only TCPM_ANALYTICS_ADDR", env: map[string]string{"TCPM_ANALYTICS_ADDR": "0.0.0.0:9000"}, want: "0.0.0.0:9000"},
		{name: "both", env: map[string]string{"TCPM_DASHBOARD_ADDR": "0.0.0.0:8080", "TCPM_ANALYTICS_ADDR": "0.0.0.0:9000"}, want: "0.0.0.0:8080"},
		{name: "a blank TCPM_DASHBOARD_ADDR", env: map[string]string{"TCPM_DASHBOARD_ADDR": "  ", "TCPM_ANALYTICS_ADDR": "0.0.0.0:9000"}, want: "0.0.0.0:9000"},
		{name: "a padded TCPM_DASHBOARD_ADDR", env: map[string]string{"TCPM_DASHBOARD_ADDR": " :8080 "}, want: ":8080"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			getenv := func(name string) string { return tt.env[name] }
			if got := dashboardAddrOverride(getenv); got != tt.want {
				t.Errorf("dashboardAddrOverride(%v) = %q, want %q", tt.env, got, tt.want)
			}
		})
	}
}

func TestDashboardAddr(t *testing.T) {
	tests := []struct {
		name     string
		cfg      dashboardConfig
		override string
		want     string
	}{
		{name: "a host and a port", cfg: dashboardConfig{Host: "0.0.0.0", Port: 8080}, want: "0.0.0.0:8080"},
		{name: "nothing set", cfg: dashboardConfig{}, want: "127.0.0.1:5000"},
		{name: "no host", cfg: dashboardConfig{Port: 8080}, want: "127.0.0.1:8080"},
		{name: "a blank host", cfg: dashboardConfig{Host: "  ", Port: 8080}, want: "127.0.0.1:8080"},
		{name: "a padded host", cfg: dashboardConfig{Host: " localhost ", Port: 8080}, want: "localhost:8080"},
		{name: "an IPv6 host", cfg: dashboardConfig{Host: "::1", Port: 8080}, want: "[::1]:8080"},
		{name: "port zero", cfg: dashboardConfig{Host: "0.0.0.0", Port: 0}, want: "0.0.0.0:5000"},
		{name: "port one", cfg: dashboardConfig{Host: "0.0.0.0", Port: 1}, want: "0.0.0.0:1"},
		{name: "a negative port", cfg: dashboardConfig{Host: "0.0.0.0", Port: -1}, want: "0.0.0.0:5000"},
		{name: "an override", cfg: dashboardConfig{Host: "0.0.0.0", Port: 8080}, override: "127.0.0.1:9000", want: "127.0.0.1:9000"},
		{name: "a padded override", cfg: dashboardConfig{Host: "0.0.0.0", Port: 8080}, override: " :9000 ", want: ":9000"},
		{name: "a blank override", cfg: dashboardConfig{Host: "0.0.0.0", Port: 8080}, override: "   ", want: "0.0.0.0:8080"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := dashboardAddr(tt.cfg, tt.override); got != tt.want {
				t.Errorf("dashboardAddr(%+v, %q) = %q, want %q", tt.cfg, tt.override, got, tt.want)
			}
		})
	}
}
