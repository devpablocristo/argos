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
	OrgID                string
	NexusBaseURL         string
	NexusAPIKey          string
	CompanionBaseURL     string
	CompanionAPIKey      string
	PublicBaseURL        string
}

func Load() Config {
	return Config{
		Port:                 envconfig.Get("PORT", "18090"),
		DatabaseURL:          envconfig.Get("ARGOS_DATABASE_URL", ""),
		StorageDir:           envconfig.Get("ARGOS_STORAGE_DIR", "../var/outputs"),
		ProcessingPython:     envconfig.Get("ARGOS_PROCESSING_PYTHON", defaultPython()),
		ProcessingPythonPath: envconfig.Get("ARGOS_PROCESSING_PYTHONPATH", "../processing/python"),
		ProcessingTimeout:    envconfig.Duration("ARGOS_PROCESSING_TIMEOUT_SEC", 120*time.Second),
		OrgID:                envconfig.Get("ARGOS_ORG_ID", "argos-local-org"),
		NexusBaseURL:         envconfig.Get("ARGOS_NEXUS_BASE_URL", ""),
		NexusAPIKey:          envconfig.Get("ARGOS_NEXUS_API_KEY", ""),
		CompanionBaseURL:     envconfig.Get("ARGOS_COMPANION_BASE_URL", ""),
		CompanionAPIKey:      envconfig.Get("ARGOS_COMPANION_API_KEY", ""),
		PublicBaseURL:        envconfig.Get("ARGOS_PUBLIC_BASE_URL", ""),
	}
}

func defaultPython() string {
	local := "../processing/python/.venv/bin/python"
	if _, err := os.Stat(local); err == nil {
		return local
	}
	return "python3"
}
