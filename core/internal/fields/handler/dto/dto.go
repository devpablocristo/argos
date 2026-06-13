package dto

type CreateFieldRequest struct {
	Name  string `json:"name"`
	Notes string `json:"notes"`
}

type UpdateFieldRequest struct {
	Name  *string `json:"name,omitempty"`
	Notes *string `json:"notes,omitempty"`
}

type FieldListResponse struct {
	Fields any `json:"fields"`
}
