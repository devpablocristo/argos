package config

import (
	"os"
	"time"

	"github.com/devpablocristo/platform/config/go/envconfig"
)

type Config struct {
	Port                 string
	DatabaseURL          string
	StorageDir           string
	ProcessingPython     string
	ProcessingPythonPath string
	ProcessingTimeout    time.Duration
}

func Load() Config {
	return Config{
		Port:                 envconfig.Get("PORT", "18090"),
		DatabaseURL:          envconfig.Get("ARGOS_DATABASE_URL", ""),
		StorageDir:           envconfig.Get("ARGOS_STORAGE_DIR", "../var/outputs"),
		ProcessingPython:     envconfig.Get("ARGOS_PROCESSING_PYTHON", defaultPython()),
		ProcessingPythonPath: envconfig.Get("ARGOS_PROCESSING_PYTHONPATH", "../processing/python"),
		ProcessingTimeout:    envconfig.Duration("ARGOS_PROCESSING_TIMEOUT_SEC", 120*time.Second),
	}
}

func defaultPython() string {
	local := "../processing/python/.venv/bin/python"
	if _, err := os.Stat(local); err == nil {
		return local
	}
	return "python3"
}
