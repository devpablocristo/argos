package wire

import (
	"context"
	"net/http"
	"time"

	"github.com/devpablocristo/argos/core/internal/analyses"
	"github.com/devpablocristo/argos/core/internal/axis"
	"github.com/devpablocristo/argos/core/internal/catalog"
	"github.com/devpablocristo/argos/core/internal/database"
	"github.com/devpablocristo/argos/core/internal/fields"
	"github.com/devpablocristo/argos/core/internal/processor"
)

type Config struct {
	DatabaseURL          string
	StorageDir           string
	ProcessingPython     string
	ProcessingPythonPath string
	ProcessingTimeoutSec int
	OrgID                string
	NexusBaseURL         string
	NexusAPIKey          string
	CompanionBaseURL     string
	CompanionAPIKey      string
	PublicBaseURL        string
}

func NewServer(cfg Config) (http.Handler, func(), error) {
	cleanup := func() {}
	var catalogRepo catalog.Repository
	var analysisRepo analyses.Repository
	var fieldsRepo fields.Repository
	if cfg.DatabaseURL != "" {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		db, err := database.Open(ctx, cfg.DatabaseURL)
		if err != nil {
			return nil, nil, err
		}
		if err := database.Migrate(ctx, db, "migrations"); err != nil {
			_ = db.Close()
			return nil, nil, err
		}
		catalogRepo = catalog.NewSQLRepository(db)
		analysisRepo = analyses.NewSQLRepository(db)
		fieldsRepo = fields.NewSQLRepository(db)
		cleanup = func() { _ = db.Close() }
	} else {
		catalogRepo = catalog.NewMemoryRepository()
		analysisRepo = analyses.NewMemoryRepository()
		fieldsRepo = fields.NewMemoryRepository()
	}
	processingWorker := processor.NewCLIProcessor(processor.CLIConfig{
		Python:     cfg.ProcessingPython,
		PythonPath: cfg.ProcessingPythonPath,
		Timeout:    durationSeconds(cfg.ProcessingTimeoutSec),
	})

	fieldsUC := fields.NewUsecases(fieldsRepo)
	catalogUC := catalog.NewUsecases(catalogRepo, processingWorker, cfg.StorageDir, fieldsUC)
	axisClient := axis.NewClient(axis.Config{
		OrgID:            cfg.OrgID,
		NexusBaseURL:     cfg.NexusBaseURL,
		NexusAPIKey:      cfg.NexusAPIKey,
		CompanionBaseURL: cfg.CompanionBaseURL,
		CompanionAPIKey:  cfg.CompanionAPIKey,
		PublicBaseURL:    cfg.PublicBaseURL,
	})
	analysisUC := analyses.NewUsecases(analysisRepo, catalogUC, axisClient)

	catalogHandler := catalog.NewHandler(catalogUC)
	fieldsHandler := fields.NewHandler(fieldsUC)
	analysisHandler := analyses.NewHandler(analysisUC)

	mux := http.NewServeMux()
	registerHealthEndpoints(mux)
	fieldsHandler.Register(mux)
	catalogHandler.Register(mux)
	analysisHandler.Register(mux)

	return withCORS(mux), cleanup, nil
}

func registerHealthEndpoints(mux *http.ServeMux) {
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), durationSeconds(1))
		defer cancel()
		select {
		case <-ctx.Done():
			writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
		default:
			writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
		}
	})
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-API-KEY")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
