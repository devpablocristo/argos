package fields

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	fielddto "github.com/devpablocristo/argos/core/internal/fields/handler/dto"
	fielddomain "github.com/devpablocristo/argos/core/internal/fields/usecases/domain"
)

type usecasePort interface {
	CreateField(ctx context.Context, name, notes string) (fielddomain.Field, error)
	ListFields(ctx context.Context, includeArchived bool) ([]fielddomain.Field, error)
	GetField(ctx context.Context, id string) (fielddomain.Field, error)
	UpdateField(ctx context.Context, id string, name, notes *string) (fielddomain.Field, error)
	ArchiveField(ctx context.Context, id string) (fielddomain.Field, error)
	RestoreField(ctx context.Context, id string) (fielddomain.Field, error)
	DeleteField(ctx context.Context, id string) (fielddomain.Field, error)
}

type Handler struct {
	uc usecasePort
}

func NewHandler(uc usecasePort) *Handler {
	return &Handler{uc: uc}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /v1/fields", h.createField)
	mux.HandleFunc("GET /v1/fields", h.listFields)
	mux.HandleFunc("GET /v1/fields/{id}", h.getField)
	mux.HandleFunc("PATCH /v1/fields/{id}", h.updateField)
	mux.HandleFunc("POST /v1/fields/{id}/archive", h.archiveField)
	mux.HandleFunc("POST /v1/fields/{id}/restore", h.restoreField)
	mux.HandleFunc("DELETE /v1/fields/{id}", h.deleteField)
}

func (h *Handler) createField(w http.ResponseWriter, r *http.Request) {
	var req fielddto.CreateFieldRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}
	field, err := h.uc.CreateField(r.Context(), req.Name, req.Notes)
	if err != nil {
		writeFieldMutationError(w, err, "create_field_failed")
		return
	}
	writeJSON(w, http.StatusCreated, field)
}

func (h *Handler) listFields(w http.ResponseWriter, r *http.Request) {
	fields, err := h.uc.ListFields(r.Context(), parseBool(r.URL.Query().Get("include_archived")))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "list_fields_failed", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, fielddto.FieldListResponse{Fields: fields})
}

func (h *Handler) getField(w http.ResponseWriter, r *http.Request) {
	field, err := h.uc.GetField(r.Context(), r.PathValue("id"))
	if err != nil {
		writeFieldMutationError(w, err, "get_field_failed")
		return
	}
	writeJSON(w, http.StatusOK, field)
}

func (h *Handler) updateField(w http.ResponseWriter, r *http.Request) {
	var req fielddto.UpdateFieldRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}
	field, err := h.uc.UpdateField(r.Context(), r.PathValue("id"), req.Name, req.Notes)
	if err != nil {
		writeFieldMutationError(w, err, "update_field_failed")
		return
	}
	writeJSON(w, http.StatusOK, field)
}

func (h *Handler) archiveField(w http.ResponseWriter, r *http.Request) {
	field, err := h.uc.ArchiveField(r.Context(), r.PathValue("id"))
	if err != nil {
		writeFieldMutationError(w, err, "archive_field_failed")
		return
	}
	writeJSON(w, http.StatusOK, field)
}

func (h *Handler) restoreField(w http.ResponseWriter, r *http.Request) {
	field, err := h.uc.RestoreField(r.Context(), r.PathValue("id"))
	if err != nil {
		writeFieldMutationError(w, err, "restore_field_failed")
		return
	}
	writeJSON(w, http.StatusOK, field)
}

func (h *Handler) deleteField(w http.ResponseWriter, r *http.Request) {
	field, err := h.uc.DeleteField(r.Context(), r.PathValue("id"))
	if err != nil {
		writeFieldMutationError(w, err, "delete_field_failed")
		return
	}
	writeJSON(w, http.StatusOK, field)
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

func writeFieldMutationError(w http.ResponseWriter, err error, code string) {
	switch {
	case errors.Is(err, ErrNotFound):
		writeError(w, http.StatusNotFound, "not_found", "field not found")
	case errors.Is(err, ErrValidation):
		writeError(w, http.StatusBadRequest, "validation_error", "field name is required")
	case errors.Is(err, ErrFieldNotArchived):
		writeError(w, http.StatusConflict, "field_not_archived", "archive field before hard delete")
	case errors.Is(err, ErrFieldHasDatasets):
		writeError(w, http.StatusConflict, "field_has_datasets", "reassign or detach datasets before hard deleting this field")
	default:
		writeError(w, http.StatusInternalServerError, code, err.Error())
	}
}

func parseBool(value string) bool {
	return value == "1" || strings.EqualFold(value, "true") || strings.EqualFold(value, "yes")
}
