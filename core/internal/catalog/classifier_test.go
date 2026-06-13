package catalog

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"
	"time"

	catalogdomain "github.com/devpablocristo/argos/core/internal/catalog/usecases/domain"
	"github.com/devpablocristo/argos/core/internal/processor"
)

func TestClassifyDataset(t *testing.T) {
	now := time.Date(2026, 5, 23, 12, 0, 0, 0, time.UTC)
	lat := -34.1
	lon := -58.2

	tests := []struct {
		name      string
		source    string
		captures  []catalogdomain.Capture
		wantType  string
		wantScope string
	}{
		{
			name:      "sample source",
			source:    "/data/sample",
			wantType:  DatasetTypeSample,
			wantScope: ScopeGlobal,
		},
		{
			name:      "uploaded folder before scan",
			source:    "/data/outputs/uploads/abc",
			wantType:  DatasetTypeUploadedFolder,
			wantScope: ScopeGlobal,
		},
		{
			name:      "single capture",
			source:    "/data/uploads/abc",
			captures:  []catalogdomain.Capture{{ID: "capture-1", CaptureKey: "capture-1"}},
			wantType:  DatasetTypeSingleCapture,
			wantScope: ScopeDataset,
		},
		{
			name:   "multiple captures without flight evidence",
			source: "/data/uploads/abc",
			captures: []catalogdomain.Capture{
				{ID: "capture-1", CaptureKey: "capture-1"},
				{ID: "capture-2", CaptureKey: "capture-2"},
			},
			wantType:  DatasetTypeMultiCaptureDataset,
			wantScope: ScopeDataset,
		},
		{
			name:   "flight dataset",
			source: "/data/uploads/abc",
			captures: []catalogdomain.Capture{
				{
					ID:         "capture-1",
					CaptureKey: "DJI_20230821133004_0001",
					CapturedAt: "2023-08-21T13:30:04Z",
					Location:   catalogdomain.Location{Lat: &lat, Lon: &lon},
				},
				{
					ID:         "capture-2",
					CaptureKey: "DJI_20230821133005_0002",
					CapturedAt: "2023-08-21T13:30:05Z",
					Location:   catalogdomain.Location{Lat: &lat, Lon: &lon},
				},
			},
			wantType:  DatasetTypeFlightDataset,
			wantScope: ScopeFlight,
		},
		{
			name:      "unknown",
			source:    "/random/path",
			wantType:  DatasetTypeUnknown,
			wantScope: ScopeGlobal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := classifyDataset(catalogdomain.Dataset{ID: "dataset-1", Name: "Dataset", SourceURI: tt.source}, tt.captures, "/data/outputs", now)
			if got.DatasetType != tt.wantType {
				t.Fatalf("DatasetType = %q, want %q", got.DatasetType, tt.wantType)
			}
			if got.Scope != tt.wantScope {
				t.Fatalf("Scope = %q, want %q", got.Scope, tt.wantScope)
			}
			if got.FieldID != nil || got.LotID != nil || got.CampaignID != nil || got.FlightID != nil {
				t.Fatalf("classification invented agronomic ids: %+v", got)
			}
			if len(got.MissingMetadata) == 0 {
				t.Fatalf("MissingMetadata is empty")
			}
		})
	}
}

func TestClassifyDatasetWithFieldAssociation(t *testing.T) {
	fieldID := "field-1"
	got := classifyDataset(catalogdomain.Dataset{
		ID:        "dataset-1",
		Name:      "Dataset",
		SourceURI: "/data/outputs/uploads/abc",
		FieldID:   &fieldID,
	}, nil, "/data/outputs", time.Now().UTC())
	if got.FieldID == nil || *got.FieldID != fieldID {
		t.Fatalf("FieldID = %v, want %q", got.FieldID, fieldID)
	}
	if got.Scope != ScopeField {
		t.Fatalf("Scope = %q, want %q", got.Scope, ScopeField)
	}
	if slices.Contains(got.MissingMetadata, "field_id") {
		t.Fatalf("MissingMetadata contains field_id: %v", got.MissingMetadata)
	}
}

func TestScanDatasetRecordsClassificationAndEvents(t *testing.T) {
	ctx := context.Background()
	repo := NewMemoryRepository()
	uc := NewUsecases(repo, fakeProcessor{
		response: processor.Response{
			Status:     "completed",
			InputPath:  "/input",
			OutputPath: "/output",
			Captures: []processor.CaptureResult{
				{
					CaptureKey: "capture-1",
					Validation: processor.Validation{Status: "valid"},
					Analysis: &processor.AnalysisResult{
						Kind:   "ndvi",
						Status: "completed",
						Outputs: []processor.OutputResult{{
							Kind:        "preview_png",
							Path:        "/output/preview.png",
							ContentType: "image/png",
						}},
						Metrics: map[string]any{"mean": 0.42},
					},
				},
			},
		},
	}, t.TempDir())

	ds, err := uc.CreateDataset(ctx, "Uploaded", "/data/outputs/uploads/abc")
	if err != nil {
		t.Fatal(err)
	}
	if ds.Classification == nil || ds.Classification.DatasetType != DatasetTypeUploadedFolder {
		t.Fatalf("initial classification = %+v, want uploaded_folder", ds.Classification)
	}
	if _, _, _, err := uc.ScanDataset(ctx, ds.ID); err != nil {
		t.Fatal(err)
	}
	classification, err := uc.GetDatasetClassification(ctx, ds.ID)
	if err != nil {
		t.Fatal(err)
	}
	if classification.DatasetType != DatasetTypeSingleCapture {
		t.Fatalf("post-scan DatasetType = %q, want %q", classification.DatasetType, DatasetTypeSingleCapture)
	}
	events, err := uc.ListDatasetEvents(ctx, ds.ID)
	if err != nil {
		t.Fatal(err)
	}
	types := make([]string, 0, len(events))
	for _, event := range events {
		types = append(types, event.EventType)
	}
	for _, want := range []string{
		"DATASET_CREATED",
		"DATASET_CLASSIFIED",
		"ANALYSIS_STARTED",
		"METADATA_EXTRACTED",
		"CAPTURES_DETECTED",
		"ANALYSIS_COMPLETED",
		"INDEX_GENERATED",
	} {
		if !slices.Contains(types, want) {
			t.Fatalf("events missing %s: %v", want, types)
		}
	}
}

func TestDatasetClassificationAndEventsEndpoints(t *testing.T) {
	ctx := context.Background()
	uc := NewUsecases(NewMemoryRepository(), fakeProcessor{}, t.TempDir())
	handler := NewHandler(uc)
	mux := http.NewServeMux()
	handler.Register(mux)

	body := bytes.NewBufferString(`{"name":"Sample","source_uri":"/data/sample"}`)
	createReq := httptest.NewRequest(http.MethodPost, "/v1/datasets", body)
	createRec := httptest.NewRecorder()
	mux.ServeHTTP(createRec, createReq)
	if createRec.Code != http.StatusCreated {
		t.Fatalf("create status = %d body=%s", createRec.Code, createRec.Body.String())
	}
	var ds catalogdomain.Dataset
	if err := json.NewDecoder(createRec.Body).Decode(&ds); err != nil {
		t.Fatal(err)
	}

	classificationReq := httptest.NewRequest(http.MethodGet, "/v1/datasets/"+ds.ID+"/classification", nil)
	classificationRec := httptest.NewRecorder()
	mux.ServeHTTP(classificationRec, classificationReq.WithContext(ctx))
	if classificationRec.Code != http.StatusOK {
		t.Fatalf("classification status = %d body=%s", classificationRec.Code, classificationRec.Body.String())
	}
	var classification catalogdomain.DatasetClassification
	if err := json.NewDecoder(classificationRec.Body).Decode(&classification); err != nil {
		t.Fatal(err)
	}
	if classification.DatasetType != DatasetTypeSample {
		t.Fatalf("DatasetType = %q, want %q", classification.DatasetType, DatasetTypeSample)
	}

	eventsReq := httptest.NewRequest(http.MethodGet, "/v1/datasets/"+ds.ID+"/events", nil)
	eventsRec := httptest.NewRecorder()
	mux.ServeHTTP(eventsRec, eventsReq.WithContext(ctx))
	if eventsRec.Code != http.StatusOK {
		t.Fatalf("events status = %d body=%s", eventsRec.Code, eventsRec.Body.String())
	}
	var eventResponse struct {
		Events []catalogdomain.DatasetEvent `json:"events"`
	}
	if err := json.NewDecoder(eventsRec.Body).Decode(&eventResponse); err != nil {
		t.Fatal(err)
	}
	if len(eventResponse.Events) < 2 {
		t.Fatalf("events length = %d, want at least 2", len(eventResponse.Events))
	}
}

type fakeProcessor struct {
	response processor.Response
	err      error
}

func (f fakeProcessor) ProcessCaptureGroups(context.Context, string, string) (processor.Response, error) {
	if f.err != nil {
		return processor.Response{}, f.err
	}
	return f.response, nil
}

func TestScanDatasetRecordsErrorEvent(t *testing.T) {
	ctx := context.Background()
	uc := NewUsecases(NewMemoryRepository(), fakeProcessor{err: errors.New("boom")}, t.TempDir())
	ds, err := uc.CreateDataset(ctx, "Broken", "/data/outputs/uploads/abc")
	if err != nil {
		t.Fatal(err)
	}
	if _, _, _, err := uc.ScanDataset(ctx, ds.ID); err == nil {
		t.Fatal("ScanDataset error = nil, want error")
	}
	events, err := uc.ListDatasetEvents(ctx, ds.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.ContainsFunc(events, func(event catalogdomain.DatasetEvent) bool {
		return event.EventType == "ERROR" && event.Status == "failed"
	}) {
		t.Fatalf("events missing failed ERROR: %+v", events)
	}
}

func TestUploadAndScanDatasetRecordsUploadEvent(t *testing.T) {
	ctx := context.Background()
	uc := NewUsecases(NewMemoryRepository(), fakeProcessor{response: processor.Response{Status: "completed"}}, t.TempDir())
	ds, _, _, _, err := uc.UploadAndScanDataset(ctx, "Upload", []UploadedFile{{
		Name: "image.jpg",
		Open: func() (io.ReadCloser, error) {
			return io.NopCloser(bytes.NewReader([]byte("jpg"))), nil
		},
	}})
	if err != nil {
		t.Fatal(err)
	}
	events, err := uc.ListDatasetEvents(ctx, ds.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.ContainsFunc(events, func(event catalogdomain.DatasetEvent) bool {
		return event.EventType == "DATASET_UPLOADED" && event.Status == "completed"
	}) {
		t.Fatalf("events missing DATASET_UPLOADED: %+v", events)
	}
}
