package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/user/portwatch/internal/config"
	"github.com/user/portwatch/internal/daemon"
)

func main() {
	cfgPath := flag.String("config", "config.yaml", "path to config file")
	snapPath := flag.String("snapshot", "/var/lib/portwatch/snapshot.json", "path to snapshot file")
	flag.Parse()

	cfg, err := config.Load(*cfgPath)
	if err != nil {
		log.Printf("could not load config (%v), using defaults", err)
		cfg = config.Default()
	}

	d, err := daemon.New(cfg, *snapPath)
	if err != nil {
		log.Fatalf("failed to initialise daemon: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := d.Run(ctx); err != nil && err != context.Canceled {
		log.Printf("daemon exited: %v", err)
	}
}
