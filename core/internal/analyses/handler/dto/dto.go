package dto

type CreateAnalysisRequest struct {
	TargetType string `json:"target_type"`
	TargetID   string `json:"target_id"`
	Kind       string `json:"kind"`
}

type OutputListResponse struct {
	Outputs any `json:"outputs"`
}
