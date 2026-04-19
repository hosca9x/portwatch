package config

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds the portwatch daemon configuration.
type Config struct {
	Interval  time.Duration `yaml:"interval"`
	AlertFile string        `yaml:"alert_file"`
	Ignore    []uint16      `yaml:"ignore"`
	Protocols []string      `yaml:"protocols"`
}

// Default returns a Config with sensible defaults.
func Default() *Config {
	return &Config{
		Interval:  30 * time.Second,
		AlertFile: "",
		Ignore:    []uint16{},
		Protocols: []string{"tcp", "udp"},
	}
}

// Load reads a YAML config file from the given path.
// Missing fields fall back to defaults.
func Load(path string) (*Config, error) {
	cfg := Default()

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	dec := yaml.NewDecoder(f)
	dec.KnownFields(true)
	if err := dec.Decode(cfg); err != nil {
		return nil, err
	}

	if len(cfg.Protocols) == 0 {
		cfg.Protocols = Default().Protocols
	}
	if cfg.Interval <= 0 {
		cfg.Interval = Default().Interval
	}

	return cfg, nil
}
