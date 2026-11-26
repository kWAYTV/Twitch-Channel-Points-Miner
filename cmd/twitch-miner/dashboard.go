package main

import (
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/kWAYTV/twitch-points-miner/internal/dashboard"
	"github.com/kWAYTV/twitch-points-miner/internal/logging"
	"github.com/kWAYTV/twitch-points-miner/internal/miner"
)

type dashboardConfig struct {
	Enabled bool   `json:"enabled"`
	Host    string `json:"host"`
	Port    int    `json:"port"`
}

func defaultDashboardConfig() map[string]interface{} {
	return map[string]interface{}{
		"enabled": true,
		"host":    "127.0.0.1",
		"port":    5000,
	}
}

// dashboardAddrOverride is the listen address the environment asks for:
// TCPM_DASHBOARD_ADDR, or TCPM_ANALYTICS_ADDR, its older name.
func dashboardAddrOverride(getenv func(string) string) string {
	if addr := strings.TrimSpace(getenv("TCPM_DASHBOARD_ADDR")); addr != "" {
		return addr
	}
	return strings.TrimSpace(getenv("TCPM_ANALYTICS_ADDR"))
}

// dashboardAddr resolves the dashboard listen address. override comes from
// the environment and wins over the config when it is set.
func dashboardAddr(cfg dashboardConfig, override string) string {
	if raw := strings.TrimSpace(override); raw != "" {
		return raw
	}
	host := strings.TrimSpace(cfg.Host)
	if host == "" {
		host = "127.0.0.1"
	}
	port := cfg.Port
	if port <= 0 {
		port = 5000
	}
	return net.JoinHostPort(host, strconv.Itoa(port))
}

// startDashboard wires the history store into the miner and serves the
// dashboard in the background. Data lives in ./analytics of the data directory.
func startDashboard(cfg dashboardConfig, logger *logging.Logger) {
	if !cfg.Enabled {
		return
	}
	store, err := dashboard.NewStore("analytics")
	if err != nil {
		logger.Errorf("dashboard disabled: %v", err)
		return
	}
	miner.SetAnalyticsSink(store)
	addr := dashboardAddr(cfg, dashboardAddrOverride(os.Getenv))
	logger.Printf("Dashboard listening on http://%s", addr)
	go func() {
		if err := dashboard.NewServer(store).ListenAndServe(addr); err != nil {
			logger.Errorf("dashboard stopped: %v", err)
		}
	}()
}
