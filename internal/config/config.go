package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Config holds portwatch runtime configuration.
type Config struct {
	Interval    int      `yaml:"interval_seconds"`
	Protocols   []string `yaml:"protocols"`
	ExcludePorts []int   `yaml:"exclude_ports"`
	SnapshotPath string  `yaml:"snapshot_path"`
	HistoryPath  string  `yaml:"history_path"`
	BaselinePath string  `yaml:"baseline_path"`
	HighPriorityPorts   []int `yaml:"high_priority_ports"`
	MediumPriorityPorts []int `yaml:"medium_priority_ports"`
	MinAlertSeverity    int   `yaml:"min_alert_severity"`
}

// Default returns a Config populated with sensible defaults.
func Default() Config {
	return Config{
		Interval:            30,
		Protocols:           []string{"tcp", "udp"},
		SnapshotPath:        "/tmp/portwatch_snapshot.json",
		HistoryPath:         "/tmp/portwatch_history.json",
		BaselinePath:        "/tmp/portwatch_baseline.json",
		HighPriorityPorts:   []int{22, 443, 3389},
		MediumPriorityPorts: []int{80, 8080, 8443},
		MinAlertSeverity:    1,
	}
}

// Load reads a YAML config file and merges it over defaults.
func Load(path string) (Config, error) {
	cfg := Default()
	data, err := os.ReadFile(path)
	if err != nil {
		return cfg, err
	}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return cfg, err
	}
	if len(cfg.Protocols) == 0 {
		cfg.Protocols = []string{"tcp", "udp"}
	}
	return cfg, nil
}
