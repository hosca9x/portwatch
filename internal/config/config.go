package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Config holds portwatch runtime configuration.
type Config struct {
	Interval     int      `yaml:"interval"`      // seconds between scans
	Protocols    []string `yaml:"protocols"`     // tcp, udp
	ExcludePorts []int    `yaml:"exclude_ports"` // ports to ignore
	SnapshotPath string   `yaml:"snapshot_path"` // where to persist state
	LogLevel     string   `yaml:"log_level"`     // info, debug, warn
}

// Default returns a Config populated with sensible defaults.
func Default() *Config {
	return &Config{
		Interval:     30,
		Protocols:    []string{"tcp", "udp"},
		ExcludePorts: []int{},
		SnapshotPath: "/tmp/portwatch_snapshot.json",
		LogLevel:     "info",
	}
}

// Load reads a YAML config file and merges it over the defaults.
func Load(path string) (*Config, error) {
	cfg := Default()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, err
	}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}
	if len(cfg.Protocols) == 0 {
		cfg.Protocols = []string{"tcp", "udp"}
	}
	return cfg, nil
}
