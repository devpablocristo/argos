package catalog

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strings"

	catalogdto "github.com/devpablocristo/argos/core/internal/catalog/handler/dto"
	catalogdomain "github.com/devpablocristo/argos/core/internal/catalog/usecases/domain"
)

type usecasePort interface {
	CreateDataset(ctx context.Context, name, sourceURI string) (catalogdomain.Dataset, error)
	ListDatasets(ctx context.Context) ([]catalogdomain.Dataset, error)
	GetDataset(ctx context.Context, id string) (catalogdomain.Dataset, error)
	ScanDataset(ctx context.Context, datasetID string) ([]catalogdomain.Capture, []string, string, error)
	ListCaptures(ctx context.Context, datasetID string) ([]catalogdomain.Capture, error)
	GetCaptureAsset(ctx context.Context, captureID, band string) (catalogdomain.BandAsset, error)
}

type Handler struct {
	uc usecasePort
}

func NewHandler(uc usecasePort) *Handler {
	return &Handler{uc: uc}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /v1/datasets", h.createDataset)
	mux.HandleFunc("GET /v1/datasets", h.listDatasets)
	mux.HandleFunc("POST /v1/datasets/{id}/scan", h.scanDataset)
	mux.HandleFunc("GET /v1/datasets/{id}/captures", h.listCaptures)
	mux.HandleFunc("GET /v1/captures/{id}/assets/{band}", h.getCaptureAsset)
}

func (h *Handler) createDataset(w http.ResponseWriter, r *http.Request) {
	var req catalogdto.CreateDatasetRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	req.SourceURI = strings.TrimSpace(req.SourceURI)
	if req.Name == "" || req.SourceURI == "" {
		writeError(w, http.StatusBadRequest, "validation_error", "name and source_uri are required")
		return
	}
	ds, err := h.uc.CreateDataset(r.Context(), req.Name, req.SourceURI)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "create_dataset_failed", err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, ds)
}

func (h *Handler) listDatasets(w http.ResponseWriter, r *http.Request) {
	datasets, err := h.uc.ListDatasets(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "list_datasets_failed", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, catalogdto.DatasetListResponse{Datasets: datasets})
}

func (h *Handler) scanDataset(w http.ResponseWriter, r *http.Request) {
	captures, warnings, status, err := h.uc.ScanDataset(r.Context(), r.PathValue("id"))
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "dataset not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "processing_failed", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, catalogdto.ScanDatasetResponse{Status: status, Captures: captures, Warnings: warnings})
}

func (h *Handler) listCaptures(w http.ResponseWriter, r *http.Request) {
	captures, err := h.uc.ListCaptures(r.Context(), r.PathValue("id"))
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "dataset not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "list_captures_failed", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, catalogdto.CaptureListResponse{Captures: captures})
}

func (h *Handler) getCaptureAsset(w http.ResponseWriter, r *http.Request) {
	asset, err := h.uc.GetCaptureAsset(r.Context(), r.PathValue("id"), r.PathValue("band"))
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "capture asset not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "get_capture_asset_failed", err.Error())
		return
	}
	if _, err := os.Stat(asset.Path); err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, os.ErrNotExist) {
			status = http.StatusNotFound
		}
		writeError(w, status, "asset_unavailable", "asset file is unavailable")
		return
	}
	w.Header().Set("Content-Type", asset.MimeType)
	http.ServeFile(w, r, asset.Path)
}

func decodeJSON(r *http.Request, dst any) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(dst)
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, map[string]any{"code": code, "message": message})
}

func datasetOutputPath(root, datasetID string) string {
	return root + "/datasets/" + datasetID
}
