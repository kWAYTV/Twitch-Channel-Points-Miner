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

type analyticsConfig struct {
	Enabled bool   `json:"enabled"`
	Host    string `json:"host"`
	Port    int    `json:"port"`
}

func defaultAnalyticsConfig() map[string]interface{} {
	return map[string]interface{}{
		"enabled": true,
		"host":    "127.0.0.1",
		"port":    5000,
	}
}

// analyticsAddr resolves the dashboard listen address; TCPM_ANALYTICS_ADDR overrides the config.
func analyticsAddr(cfg analyticsConfig) string {
	if raw := strings.TrimSpace(os.Getenv("TCPM_ANALYTICS_ADDR")); raw != "" {
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

// startAnalytics wires the analytics store into the miner and serves the
// dashboard in the background. Data lives in ./analytics of the data directory.
func startAnalytics(cfg analyticsConfig, logger *logging.Logger) {
	if !cfg.Enabled {
		return
	}
	store, err := dashboard.NewStore("analytics")
	if err != nil {
		logger.Errorf("analytics disabled: %v", err)
		return
	}
	miner.SetAnalyticsSink(store)
	addr := analyticsAddr(cfg)
	logger.Printf("Analytics dashboard listening on http://%s", addr)
	go func() {
		if err := dashboard.NewServer(store).ListenAndServe(addr); err != nil {
			logger.Errorf("analytics dashboard stopped: %v", err)
		}
	}()
}
