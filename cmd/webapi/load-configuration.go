package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/ardanlabs/conf"
	"gopkg.in/yaml.v2"
)

// WebAPIConfiguration describes the web API configuration
type WebAPIConfiguration struct {
	Config struct {
		Path string `conf:"default:data/config.yaml"` // Config file
	}
	Web struct {
		APIHost         string        `conf:"default:0.0.0.0:3000"`
		DebugHost       string        `conf:"default:0.0.0.0:4000"`
		ReadTimeout     time.Duration `conf:"default:10s"`
		WriteTimeout    time.Duration `conf:"default:10s"`
		ShutdownTimeout time.Duration `conf:"default:10s"`
	}
	Debug bool
	DB    struct {
		Filename string `conf:"default:data/wasatext_2.db"` // SQLite DB path
	}
}

// loadConfiguration reads CLI flags, env vars, then YAML config
func loadConfiguration() (WebAPIConfiguration, error) {
	var cfg WebAPIConfiguration

	// Parse CLI and env
	if err := conf.Parse(os.Args[1:], "CFG", &cfg); err != nil {
		if errors.Is(err, conf.ErrHelpWanted) {
			usage, err := conf.Usage("CFG", &cfg)
			if err != nil {
				return cfg, fmt.Errorf("generating config usage: %w", err)
			}
			fmt.Println(usage)
			return cfg, conf.ErrHelpWanted
		}
		return cfg, fmt.Errorf("parsing config: %w", err)
	}

	// Load YAML config file if exists
	fp, err := os.Open(cfg.Config.Path)
	if err != nil && !os.IsNotExist(err) {
		return cfg, fmt.Errorf("cannot read config file: %w", err)
	} else if err == nil {
		defer fp.Close()
		yamlFile, err := io.ReadAll(fp)
		if err != nil {
			return cfg, fmt.Errorf("cannot read YAML: %w", err)
		}
		if err := yaml.Unmarshal(yamlFile, &cfg); err != nil {
			return cfg, fmt.Errorf("cannot unmarshal YAML: %w", err)
		}
	}

	// Ensure data folder exists
	dbDir := filepath.Dir(cfg.DB.Filename)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return cfg, fmt.Errorf("cannot create data folder: %w", err)
	}

	return cfg, nil
}
