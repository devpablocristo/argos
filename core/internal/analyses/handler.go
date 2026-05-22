package analyses

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"

	analysisdto "github.com/devpablocristo/argos/core/internal/analyses/handler/dto"
	analysisdomain "github.com/devpablocristo/argos/core/internal/analyses/usecases/domain"
)

type usecasePort interface {
	CreateNDVIForCapture(ctx context.Context, captureID string) (analysisdomain.Analysis, error)
	GetAnalysis(ctx context.Context, id string) (analysisdomain.Analysis, error)
	ListAnalysisOutputs(ctx context.Context, analysisID string) ([]analysisdomain.OutputAsset, error)
	GetOutput(ctx context.Context, id string) (analysisdomain.OutputAsset, error)
}

type Handler struct {
	uc usecasePort
}

func NewHandler(uc usecasePort) *Handler {
	return &Handler{uc: uc}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /v1/analyses", h.createAnalysis)
	mux.HandleFunc("GET /v1/analyses/{id}", h.getAnalysis)
	mux.HandleFunc("GET /v1/analyses/{id}/outputs", h.listOutputs)
	mux.HandleFunc("GET /v1/assets/{id}", h.getAsset)
}

func (h *Handler) createAnalysis(w http.ResponseWriter, r *http.Request) {
	var req analysisdto.CreateAnalysisRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}
	if req.TargetType != "capture" || req.Kind != "ndvi" {
		writeError(w, http.StatusBadRequest, "unsupported_analysis", "only target_type=capture and kind=ndvi are supported")
		return
	}
	analysis, err := h.uc.CreateNDVIForCapture(r.Context(), req.TargetID)
	if err != nil {
		switch {
		case errors.Is(err, ErrNotFound):
			writeError(w, http.StatusNotFound, "not_found", "capture not found")
		case errors.Is(err, ErrAnalysisNotAvailable):
			writeError(w, http.StatusConflict, "analysis_not_available", "capture has no processing result; scan dataset first")
		default:
			writeError(w, http.StatusInternalServerError, "create_analysis_failed", err.Error())
		}
		return
	}
	writeJSON(w, http.StatusCreated, analysis)
}

func (h *Handler) getAnalysis(w http.ResponseWriter, r *http.Request) {
	analysis, err := h.uc.GetAnalysis(r.Context(), r.PathValue("id"))
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "analysis not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "get_analysis_failed", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, analysis)
}

func (h *Handler) listOutputs(w http.ResponseWriter, r *http.Request) {
	outputs, err := h.uc.ListAnalysisOutputs(r.Context(), r.PathValue("id"))
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "analysis not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "list_outputs_failed", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, analysisdto.OutputListResponse{Outputs: outputs})
}

func (h *Handler) getAsset(w http.ResponseWriter, r *http.Request) {
	output, err := h.uc.GetOutput(r.Context(), r.PathValue("id"))
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "asset not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "get_asset_failed", err.Error())
		return
	}
	if _, err := os.Stat(output.Path); err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, os.ErrNotExist) {
			status = http.StatusNotFound
		}
		writeError(w, status, "asset_unavailable", "asset file is unavailable")
		return
	}
	w.Header().Set("Content-Type", output.ContentType)
	http.ServeFile(w, r, output.Path)
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
