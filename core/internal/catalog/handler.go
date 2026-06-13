package catalog

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strings"

	catalogdto "github.com/devpablocristo/argos/core/internal/catalog/handler/dto"
	catalogdomain "github.com/devpablocristo/argos/core/internal/catalog/usecases/domain"
)

type usecasePort interface {
	CreateDataset(ctx context.Context, name, sourceURI string) (catalogdomain.Dataset, error)
	CreateDatasetForField(ctx context.Context, fieldID, name, sourceURI string) (catalogdomain.Dataset, error)
	UploadAndScanDataset(ctx context.Context, name string, files []UploadedFile) (catalogdomain.Dataset, []catalogdomain.Capture, []string, string, error)
	UploadAndScanDatasetForField(ctx context.Context, fieldID, name string, files []UploadedFile) (catalogdomain.Dataset, []catalogdomain.Capture, []string, string, error)
	ListDatasets(ctx context.Context, includeArchived bool) ([]catalogdomain.Dataset, error)
	ListDatasetsByField(ctx context.Context, fieldID string, includeArchived bool) ([]catalogdomain.Dataset, error)
	GetDataset(ctx context.Context, id string) (catalogdomain.Dataset, error)
	UpdateDataset(ctx context.Context, id string, name, sourceURI *string) (catalogdomain.Dataset, error)
	UpdateDatasetField(ctx context.Context, id string, fieldID *string) (catalogdomain.Dataset, error)
	ArchiveDataset(ctx context.Context, id string) (catalogdomain.Dataset, error)
	RestoreDataset(ctx context.Context, id string) (catalogdomain.Dataset, error)
	DeleteDataset(ctx context.Context, id string) (catalogdomain.Dataset, error)
	ScanDataset(ctx context.Context, datasetID string) ([]catalogdomain.Capture, []string, string, error)
	ClassifyDataset(ctx context.Context, datasetID string) (catalogdomain.DatasetClassification, error)
	GetDatasetClassification(ctx context.Context, datasetID string) (catalogdomain.DatasetClassification, error)
	ListDatasetEvents(ctx context.Context, datasetID string) ([]catalogdomain.DatasetEvent, error)
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
	mux.HandleFunc("POST /v1/datasets/upload-scan", h.uploadAndScanDataset)
	mux.HandleFunc("GET /v1/datasets", h.listDatasets)
	mux.HandleFunc("GET /v1/datasets/{id}", h.getDataset)
	mux.HandleFunc("PATCH /v1/datasets/{id}", h.updateDataset)
	mux.HandleFunc("POST /v1/datasets/{id}/archive", h.archiveDataset)
	mux.HandleFunc("POST /v1/datasets/{id}/restore", h.restoreDataset)
	mux.HandleFunc("DELETE /v1/datasets/{id}", h.deleteDataset)
	mux.HandleFunc("POST /v1/datasets/{id}/scan", h.scanDataset)
	mux.HandleFunc("PATCH /v1/datasets/{id}/field", h.updateDatasetField)
	mux.HandleFunc("GET /v1/datasets/{id}/classification", h.getDatasetClassification)
	mux.HandleFunc("POST /v1/datasets/{id}/classify", h.classifyDataset)
	mux.HandleFunc("GET /v1/datasets/{id}/events", h.listDatasetEvents)
	mux.HandleFunc("GET /v1/datasets/{id}/captures", h.listCaptures)
	mux.HandleFunc("POST /v1/fields/{id}/datasets", h.createFieldDataset)
	mux.HandleFunc("POST /v1/fields/{id}/datasets/upload-scan", h.uploadAndScanFieldDataset)
	mux.HandleFunc("GET /v1/fields/{id}/datasets", h.listFieldDatasets)
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

func (h *Handler) uploadAndScanDataset(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadBytes)
	if err := r.ParseMultipartForm(64 << 20); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid multipart upload")
		return
	}
	files := collectUploadedFiles(r.MultipartForm)
	if len(files) == 0 {
		writeError(w, http.StatusBadRequest, "validation_error", "select at least one image file")
		return
	}
	ds, captures, warnings, status, err := h.uc.UploadAndScanDataset(r.Context(), r.FormValue("name"), files)
	if err != nil {
		switch {
		case errors.Is(err, ErrValidation):
			writeError(w, http.StatusBadRequest, "validation_error", "upload must contain jpg, jpeg, tif, tiff or png image files")
		default:
			writeError(w, http.StatusInternalServerError, "upload_scan_failed", err.Error())
		}
		return
	}
	writeJSON(w, http.StatusCreated, catalogdto.UploadScanDatasetResponse{Dataset: ds, Status: status, Captures: captures, Warnings: warnings})
}

func (h *Handler) createFieldDataset(w http.ResponseWriter, r *http.Request) {
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
	ds, err := h.uc.CreateDatasetForField(r.Context(), r.PathValue("id"), req.Name, req.SourceURI)
	if err != nil {
		writeDatasetMutationError(w, err, "create_dataset_failed")
		return
	}
	writeJSON(w, http.StatusCreated, ds)
}

func (h *Handler) uploadAndScanFieldDataset(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadBytes)
	if err := r.ParseMultipartForm(64 << 20); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid multipart upload")
		return
	}
	files := collectUploadedFiles(r.MultipartForm)
	if len(files) == 0 {
		writeError(w, http.StatusBadRequest, "validation_error", "select at least one image file")
		return
	}
	ds, captures, warnings, status, err := h.uc.UploadAndScanDatasetForField(r.Context(), r.PathValue("id"), r.FormValue("name"), files)
	if err != nil {
		switch {
		case errors.Is(err, ErrValidation):
			writeError(w, http.StatusBadRequest, "validation_error", "upload must contain jpg, jpeg, tif, tiff or png image files")
		default:
			writeDatasetMutationError(w, err, "upload_scan_failed")
		}
		return
	}
	writeJSON(w, http.StatusCreated, catalogdto.UploadScanDatasetResponse{Dataset: ds, Status: status, Captures: captures, Warnings: warnings})
}

func (h *Handler) listDatasets(w http.ResponseWriter, r *http.Request) {
	datasets, err := h.uc.ListDatasets(r.Context(), parseBool(r.URL.Query().Get("include_archived")))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "list_datasets_failed", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, catalogdto.DatasetListResponse{Datasets: datasets})
}

func (h *Handler) listFieldDatasets(w http.ResponseWriter, r *http.Request) {
	datasets, err := h.uc.ListDatasetsByField(r.Context(), r.PathValue("id"), parseBool(r.URL.Query().Get("include_archived")))
	if err != nil {
		writeDatasetMutationError(w, err, "list_datasets_failed")
		return
	}
	writeJSON(w, http.StatusOK, catalogdto.DatasetListResponse{Datasets: datasets})
}

func (h *Handler) getDataset(w http.ResponseWriter, r *http.Request) {
	ds, err := h.uc.GetDataset(r.Context(), r.PathValue("id"))
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "dataset not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "get_dataset_failed", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, ds)
}

func (h *Handler) updateDataset(w http.ResponseWriter, r *http.Request) {
	var req catalogdto.UpdateDatasetRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}
	ds, err := h.uc.UpdateDataset(r.Context(), r.PathValue("id"), req.Name, req.SourceURI)
	if err != nil {
		switch {
		case errors.Is(err, ErrNotFound):
			writeError(w, http.StatusNotFound, "not_found", "dataset not found")
		case errors.Is(err, ErrValidation):
			writeError(w, http.StatusBadRequest, "validation_error", "name and source_uri cannot be empty")
		default:
			writeError(w, http.StatusInternalServerError, "update_dataset_failed", err.Error())
		}
		return
	}
	writeJSON(w, http.StatusOK, ds)
}

func (h *Handler) updateDatasetField(w http.ResponseWriter, r *http.Request) {
	var req catalogdto.UpdateDatasetFieldRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}
	ds, err := h.uc.UpdateDatasetField(r.Context(), r.PathValue("id"), req.FieldID)
	if err != nil {
		writeDatasetMutationError(w, err, "update_dataset_field_failed")
		return
	}
	writeJSON(w, http.StatusOK, ds)
}

func (h *Handler) archiveDataset(w http.ResponseWriter, r *http.Request) {
	ds, err := h.uc.ArchiveDataset(r.Context(), r.PathValue("id"))
	if err != nil {
		writeDatasetMutationError(w, err, "archive_dataset_failed")
		return
	}
	writeJSON(w, http.StatusOK, ds)
}

func (h *Handler) restoreDataset(w http.ResponseWriter, r *http.Request) {
	ds, err := h.uc.RestoreDataset(r.Context(), r.PathValue("id"))
	if err != nil {
		writeDatasetMutationError(w, err, "restore_dataset_failed")
		return
	}
	writeJSON(w, http.StatusOK, ds)
}

func (h *Handler) deleteDataset(w http.ResponseWriter, r *http.Request) {
	ds, err := h.uc.DeleteDataset(r.Context(), r.PathValue("id"))
	if err != nil {
		writeDatasetMutationError(w, err, "delete_dataset_failed")
		return
	}
	writeJSON(w, http.StatusOK, ds)
}

func (h *Handler) scanDataset(w http.ResponseWriter, r *http.Request) {
	captures, warnings, status, err := h.uc.ScanDataset(r.Context(), r.PathValue("id"))
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "dataset not found")
			return
		}
		if errors.Is(err, ErrDatasetArchived) {
			writeError(w, http.StatusConflict, "dataset_archived", "restore dataset before scanning")
			return
		}
		writeError(w, http.StatusInternalServerError, "processing_failed", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, catalogdto.ScanDatasetResponse{Status: status, Captures: captures, Warnings: warnings})
}

func (h *Handler) getDatasetClassification(w http.ResponseWriter, r *http.Request) {
	classification, err := h.uc.GetDatasetClassification(r.Context(), r.PathValue("id"))
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "dataset not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "get_dataset_classification_failed", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, classification)
}

func (h *Handler) classifyDataset(w http.ResponseWriter, r *http.Request) {
	classification, err := h.uc.ClassifyDataset(r.Context(), r.PathValue("id"))
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "dataset not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "classify_dataset_failed", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, classification)
}

func (h *Handler) listDatasetEvents(w http.ResponseWriter, r *http.Request) {
	events, err := h.uc.ListDatasetEvents(r.Context(), r.PathValue("id"))
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "dataset not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "list_dataset_events_failed", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, catalogdto.DatasetEventListResponse{Events: events})
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

func writeDatasetMutationError(w http.ResponseWriter, err error, code string) {
	if errors.Is(err, ErrNotFound) {
		writeError(w, http.StatusNotFound, "not_found", "dataset not found")
		return
	}
	if errors.Is(err, ErrDatasetNotArchived) {
		writeError(w, http.StatusConflict, "dataset_not_archived", "archive dataset before hard delete")
		return
	}
	if errors.Is(err, ErrValidation) {
		writeError(w, http.StatusBadRequest, "validation_error", "invalid dataset request")
		return
	}
	writeError(w, http.StatusInternalServerError, code, err.Error())
}

func parseBool(value string) bool {
	return value == "1" || strings.EqualFold(value, "true") || strings.EqualFold(value, "yes")
}

const maxUploadBytes int64 = 4 << 30

func collectUploadedFiles(form *multipart.Form) []UploadedFile {
	if form == nil {
		return nil
	}
	headers := form.File["files"]
	if len(headers) == 0 {
		for _, items := range form.File {
			headers = append(headers, items...)
		}
	}
	out := make([]UploadedFile, 0, len(headers))
	for _, header := range headers {
		h := header
		out = append(out, UploadedFile{
			Name: h.Filename,
			Open: func() (io.ReadCloser, error) {
				return h.Open()
			},
		})
	}
	return out
}

func datasetOutputPath(root, datasetID string) string {
	return root + "/datasets/" + datasetID
}
