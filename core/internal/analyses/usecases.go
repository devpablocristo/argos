package analyses

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	analysisdomain "github.com/devpablocristo/argos/core/internal/analyses/usecases/domain"
	"github.com/devpablocristo/argos/core/internal/axis"
	"github.com/devpablocristo/argos/core/internal/catalog"
	catalogdomain "github.com/devpablocristo/argos/core/internal/catalog/usecases/domain"
	"github.com/google/uuid"
)

type CaptureGetter interface {
	GetCapture(ctx context.Context, id string) (catalogdomain.Capture, error)
}

type DatasetEventRecorder interface {
	RecordDatasetEvent(ctx context.Context, datasetID, eventType, status, message string, details map[string]any) error
}

type Usecases struct {
	repo       Repository
	catalog    CaptureGetter
	axisClient *axis.Client
}

func NewUsecases(repo Repository, catalog CaptureGetter, axisClient ...*axis.Client) *Usecases {
	var client *axis.Client
	if len(axisClient) > 0 {
		client = axisClient[0]
	}
	return &Usecases{repo: repo, catalog: catalog, axisClient: client}
}

func (u *Usecases) CreateNDVIForCapture(ctx context.Context, captureID string) (analysisdomain.Analysis, error) {
	capture, err := u.catalog.GetCapture(ctx, captureID)
	if err != nil {
		if errors.Is(err, catalog.ErrNotFound) {
			return analysisdomain.Analysis{}, ErrNotFound
		}
		return analysisdomain.Analysis{}, err
	}
	if capture.Analysis == nil {
		return analysisdomain.Analysis{}, ErrAnalysisNotAvailable
	}
	existing, err := u.repo.GetLatestByCaptureKind(ctx, captureID, "ndvi")
	if err == nil {
		return u.ensureInsights(ctx, existing)
	}
	if err != nil && !errors.Is(err, ErrNotFound) {
		return analysisdomain.Analysis{}, err
	}
	now := time.Now().UTC()
	analysis := analysisdomain.Analysis{
		ID:          uuid.NewString(),
		DatasetID:   capture.DatasetID,
		CaptureID:   capture.ID,
		Kind:        capture.Analysis.Kind,
		Status:      capture.Analysis.Status,
		Metrics:     capture.Analysis.Metrics,
		Warnings:    append([]string(nil), capture.Analysis.Warnings...),
		CreatedAt:   now,
		CompletedAt: &now,
	}
	for _, output := range capture.Analysis.Outputs {
		analysis.Outputs = append(analysis.Outputs, analysisdomain.OutputAsset{
			ID:          uuid.NewString(),
			AnalysisID:  analysis.ID,
			Kind:        output.Kind,
			Path:        output.Path,
			ContentType: output.ContentType,
			ByteSize:    output.ByteSize,
			Metadata:    output.Metadata,
		})
	}
	u.recordDatasetEvent(ctx, capture.DatasetID, "ANALYSIS_STARTED", "running", "Persisting capture analysis in Argos.", map[string]any{
		"capture_id": capture.ID,
		"kind":       capture.Analysis.Kind,
	})
	created, err := u.repo.CreateAnalysis(ctx, analysis)
	if err != nil {
		u.recordDatasetEvent(ctx, capture.DatasetID, "ERROR", "failed", "Persisting capture analysis failed.", map[string]any{
			"capture_id": capture.ID,
			"kind":       capture.Analysis.Kind,
			"error":      err.Error(),
		})
		return analysisdomain.Analysis{}, err
	}
	u.recordDatasetEvent(ctx, capture.DatasetID, "ANALYSIS_COMPLETED", "completed", "Capture analysis persisted in Argos.", map[string]any{
		"analysis_id": created.ID,
		"capture_id":  capture.ID,
		"kind":        created.Kind,
	})
	for _, output := range created.Outputs {
		u.recordDatasetEvent(ctx, capture.DatasetID, "INDEX_GENERATED", "completed", "Analysis output registered in Argos.", map[string]any{
			"analysis_id": created.ID,
			"capture_id":  capture.ID,
			"kind":        output.Kind,
			"asset_id":    output.ID,
		})
	}
	return u.ensureInsights(ctx, created)
}

func (u *Usecases) GetAnalysis(ctx context.Context, id string) (analysisdomain.Analysis, error) {
	return u.repo.GetAnalysis(ctx, id)
}

func (u *Usecases) ListAnalysisOutputs(ctx context.Context, analysisID string) ([]analysisdomain.OutputAsset, error) {
	if _, err := u.repo.GetAnalysis(ctx, analysisID); err != nil {
		return nil, err
	}
	return u.repo.ListAnalysisOutputs(ctx, analysisID)
}

func (u *Usecases) GetOutput(ctx context.Context, id string) (analysisdomain.OutputAsset, error) {
	return u.repo.GetOutput(ctx, id)
}

var ErrAnalysisNotAvailable = errors.New("analysis not available")

func (u *Usecases) ensureInsights(ctx context.Context, analysis analysisdomain.Analysis) (analysisdomain.Analysis, error) {
	if analysis.ID == "" {
		return analysis, nil
	}
	if analysis.NexusSyncStatus == "" || analysis.NexusSyncStatus == "pending" || analysis.NexusSyncStatus == "failed" {
		analysis = u.syncNexus(ctx, analysis)
	}
	if analysis.CompanionSyncStatus == "" || analysis.CompanionSyncStatus == "pending" || analysis.CompanionSyncStatus == "failed" {
		analysis = u.syncCompanion(ctx, analysis)
	}
	updated, err := u.repo.GetAnalysis(ctx, analysis.ID)
	if err != nil {
		return analysis, nil
	}
	return updated, nil
}

func (u *Usecases) syncNexus(ctx context.Context, analysis analysisdomain.Analysis) analysisdomain.Analysis {
	if u.axisClient == nil || !u.axisClient.NexusConfigured() {
		_ = u.repo.UpdateNexusSnapshot(ctx, analysis.ID, "not_configured", "", "", analysis.NexusFindings)
		analysis.NexusSyncStatus = "not_configured"
		return analysis
	}
	facts := analysisFacts(analysis, u.axisClient.OrgID())
	resp, err := u.axisClient.SubmitFacts(ctx, axis.FactEvaluationRequest{
		OwnerSystem:   "argos",
		SourceSystem:  "argos",
		FactType:      "argos.analysis_facts.v1",
		SourceEventID: analysis.ID,
		SubjectType:   "capture",
		SubjectID:     analysis.CaptureID,
		Facts:         facts,
	})
	if err != nil {
		_ = u.repo.UpdateNexusSnapshot(ctx, analysis.ID, "failed", "", err.Error(), analysis.NexusFindings)
		analysis.NexusSyncStatus = "failed"
		analysis.NexusSyncError = err.Error()
		return analysis
	}
	findings := axis.FindingsSnapshot(resp.Findings)
	_ = u.repo.UpdateNexusSnapshot(ctx, analysis.ID, "synced", resp.Evaluation.ID, "", findings)
	analysis.NexusSyncStatus = "synced"
	analysis.NexusSyncError = ""
	analysis.NexusCorrelationID = resp.Evaluation.ID
	analysis.NexusFindings = findings
	return analysis
}

func (u *Usecases) syncCompanion(ctx context.Context, analysis analysisdomain.Analysis) analysisdomain.Analysis {
	if u.axisClient == nil || !u.axisClient.CompanionConfigured() {
		_ = u.repo.UpdateCompanionSnapshot(ctx, analysis.ID, "not_configured", "", "", analysis.CompanionOutput)
		analysis.CompanionSyncStatus = "not_configured"
		return analysis
	}
	facts := analysisFacts(analysis, u.axisClient.OrgID())
	input := map[string]any{
		"schema_version":  "argos.analysis_assist_input.v1",
		"analysis_facts":  facts,
		"findings":        analysis.NexusFindings,
		"nexus_status":    analysis.NexusSyncStatus,
		"nexus_error":     analysis.NexusSyncError,
		"analysis_public": analysisPublicURL(u.axisClient.PublicBaseURL(), analysis.ID),
	}
	resp, err := u.axisClient.RunAssist(ctx, axis.AssistRunRequest{
		OwnerSystem:    "argos",
		ProductSurface: "argos",
		AssistType:     "argos.analysis_explanation.v1",
		SubjectType:    "analysis",
		SubjectID:      analysis.ID,
		Input:          input,
	})
	if err != nil {
		output := analysis.CompanionOutput
		if len(resp.Output) > 0 {
			output = resp.Output
		}
		_ = u.repo.UpdateCompanionSnapshot(ctx, analysis.ID, "failed", resp.ID, err.Error(), output)
		analysis.CompanionSyncStatus = "failed"
		analysis.CompanionSyncError = err.Error()
		analysis.CompanionCorrelationID = resp.ID
		analysis.CompanionOutput = output
		return analysis
	}
	_ = u.repo.UpdateCompanionSnapshot(ctx, analysis.ID, "synced", resp.ID, "", resp.Output)
	analysis.CompanionSyncStatus = "synced"
	analysis.CompanionSyncError = ""
	analysis.CompanionCorrelationID = resp.ID
	analysis.CompanionOutput = resp.Output
	return analysis
}

func analysisFacts(analysis analysisdomain.Analysis, orgID string) map[string]any {
	summary := map[string]any{
		"mean":                   metricFloat(analysis.Metrics, "mean"),
		"min":                    metricFloat(analysis.Metrics, "min"),
		"max":                    metricFloat(analysis.Metrics, "max"),
		"std":                    metricFloat(analysis.Metrics, "std"),
		"valid_pixels":           metricFloat(analysis.Metrics, "valid_pixels"),
		"non_vegetation_percent": metricFloat(analysis.Metrics, "non_vegetation_percent"),
		"low_vigor_percent":      metricFloat(analysis.Metrics, "low_vigor_percent"),
		"medium_vigor_percent":   metricFloat(analysis.Metrics, "medium_vigor_percent"),
		"high_vigor_percent":     metricFloat(analysis.Metrics, "high_vigor_percent"),
	}
	return map[string]any{
		"schema_version":         "argos.analysis_facts.v1",
		"org_id":                 orgID,
		"analysis_id":            analysis.ID,
		"dataset_id":             analysis.DatasetID,
		"capture_id":             analysis.CaptureID,
		"kind":                   analysis.Kind,
		"status":                 analysis.Status,
		"metrics":                analysis.Metrics,
		"summary":                summary,
		"warnings":               analysis.Warnings,
		"mean":                   summary["mean"],
		"min":                    summary["min"],
		"max":                    summary["max"],
		"std":                    summary["std"],
		"valid_pixels":           summary["valid_pixels"],
		"non_vegetation_percent": summary["non_vegetation_percent"],
		"low_vigor_percent":      summary["low_vigor_percent"],
		"medium_vigor_percent":   summary["medium_vigor_percent"],
		"high_vigor_percent":     summary["high_vigor_percent"],
	}
}

func metricFloat(metrics map[string]any, key string) float64 {
	if metrics == nil {
		return 0
	}
	switch value := metrics[key].(type) {
	case float64:
		return value
	case float32:
		return float64(value)
	case int:
		return float64(value)
	case int64:
		return float64(value)
	case jsonNumber:
		f, _ := value.Float64()
		return f
	default:
		return 0
	}
}

type jsonNumber interface {
	Float64() (float64, error)
}

func analysisPublicURL(baseURL, analysisID string) string {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if baseURL == "" || analysisID == "" {
		return ""
	}
	return fmt.Sprintf("%s/analyses/%s", baseURL, analysisID)
}

func (u *Usecases) recordDatasetEvent(ctx context.Context, datasetID, eventType, status, message string, details map[string]any) {
	recorder, ok := u.catalog.(DatasetEventRecorder)
	if !ok || datasetID == "" {
		return
	}
	_ = recorder.RecordDatasetEvent(ctx, datasetID, eventType, status, message, details)
}
