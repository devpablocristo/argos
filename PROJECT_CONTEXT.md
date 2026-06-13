# Argos Project Context

This file gives project context to humans and agents entering the Argos repo.
Keep it synchronized with the code and documentation, and prefer incremental
updates over rewrites.

## Purpose

Argos is an agricultural analysis product for drone and multispectral imagery.
The current implementation focuses on DJI Mavic 3 Multispectral capture groups
with RGB plus green, red, red-edge and NIR bands.

The first working product path is:

1. Create or choose a field.
2. Register or upload a dataset under that field.
3. Scan source files.
4. Detect DJI capture groups.
5. Validate RGB plus multispectral bands.
6. Generate NDVI.
7. Store outputs and metadata.
8. Show the result in the UI.
9. Optionally sync facts/findings/assisted interpretation with Axis services.

## Repository Shape

- `core`: Go backend. Owns HTTP API, orchestration, repositories, migrations,
  persistence and integration boundaries.
- `processing`: Python package and CLI. Owns image inspection, metadata
  extraction, capture group discovery, validation, NDVI and raster exports.
- `ui`: React + TypeScript frontend.
- `docs`: architecture, roadmap, contracts and ADRs.
- `sample`: local DJI sample capture group.

The Python runtime is invoked by `core` through a JSON CLI contract. It is not
an HTTP service in the MVP.

## Domain Model

The professional domain hierarchy Argos is preparing for is:

```text
Field / Establishment
  Lot / Plot
    Campaign
      Flight / Mission
        Dataset
          Capture
            Analysis
```

Current important rule: a dataset is not necessarily a field, lot, campaign,
flight or capture. A dataset is the raw-file ingestion unit registered or
uploaded into Argos.

Concept meanings:

- Field / Establishment: productive physical place.
- Lot / Plot: productive subdivision within a field.
- Campaign: agricultural cycle or production period.
- Flight / Mission: drone operation at a date/time; may contain one or more
  captures.
- Dataset: raw files entered into Argos. It may be a sample, upload folder,
  flight dataset, one capture, multiple captures, sector capture or unknown.
- Capture: coherent group of images/bands from one scene or take.
- Analysis: generated result such as NDVI, RGB preview, maps, findings or
  assisted interpretation.

## Current v1 State

Formal field CRUD exists in v1 so a user can group datasets under a productive
place before loading imagery. Formal lot, campaign and flight CRUD does not
exist yet. Argos still stores nullable semantic references so future versions
can filter or migrate safely.

Current persisted entities include:

- `fields`
- `datasets`
- `flights`
- `captures`
- `band_assets`
- `analyses`
- `analysis_outputs`
- `dataset_classifications`
- `dataset_events`

`dataset_classifications` stores:

- `dataset_type`: `sample`, `uploaded_folder`, `flight_dataset`,
  `single_capture`, `multi_capture_dataset`, `sector_capture`, or `unknown`.
- `scope`: `global`, `field`, `lot`, `campaign`, `flight`, or `dataset`.
- nullable field, lot, campaign and flight references.
- confidence, missing metadata, reason and classification time.

`dataset_events` stores append-only dataset history events such as:

- `DATASET_CREATED`
- `DATASET_UPLOADED`
- `DATASET_CLASSIFIED`
- `METADATA_EXTRACTED`
- `CAPTURES_DETECTED`
- `ANALYSIS_STARTED`
- `ANALYSIS_COMPLETED`
- `INDEX_GENERATED`
- `REPORT_GENERATED`
- `ERROR`

The v1 UI starts from fields, then shows datasets associated with the selected
field. A global dataset history may still exist as a technical view, but it must
not be presented as field history unless a dataset has a real `field_id`.

## Processing and Outputs

The current processing milestone supports:

- DJI M3M capture group discovery.
- RGB plus green, red, red-edge and NIR validation.
- NDVI calculation: `(NIR - RED) / (NIR + RED)`.
- Preview PNG export.
- Analysis TIFF export.
- Metadata JSON export.

Argos does not invent georeferencing. If a source raster exposes CRS/transform,
future exports may preserve it. If only GPS metadata is available, the result is
metadata-only context.

## API and Contracts

Important internal contracts:

- `argos.eye.capture_group.v1`: produced by `processing`, consumed by `core`.
- `argos.analysis_result.v1`: produced by `processing` after NDVI export.
- `argos.analysis_facts.v1`: facts Argos sends to Nexus for deterministic
  finding evaluation.

Important dataset endpoints include:

- `GET /v1/datasets`
- `POST /v1/datasets`
- `POST /v1/datasets/upload-scan`
- `GET /v1/datasets/{id}`
- `PATCH /v1/datasets/{id}`
- `POST /v1/datasets/{id}/scan`
- `GET /v1/datasets/{id}/captures`
- `GET /v1/datasets/{id}/classification`
- `POST /v1/datasets/{id}/classify`
- `GET /v1/datasets/{id}/events`
- `PATCH /v1/datasets/{id}/field`
- `POST /v1/fields/{id}/datasets`
- `POST /v1/fields/{id}/datasets/upload-scan`
- `GET /v1/fields/{id}/datasets`

Important field endpoints include:

- `GET /v1/fields`
- `POST /v1/fields`
- `GET /v1/fields/{id}`
- `PATCH /v1/fields/{id}`
- `POST /v1/fields/{id}/archive`
- `POST /v1/fields/{id}/restore`
- `DELETE /v1/fields/{id}`

Important analysis endpoints include:

- `POST /v1/analyses`
- `GET /v1/analyses/{id}`
- `GET /v1/analyses/{id}/outputs`
- `GET /v1/assets/{id}`

## Axis Integration

Argos can work with Nexus and Companion through Axis local networking.

- Argos owns agricultural analysis facts and product-owned seeds.
- Nexus evaluates deterministic finding rules and stores findings.
- Companion generates assisted AI interpretation from facts and findings.
- Nexus should not depend on Companion.
- Companion may consume Nexus context.

Local defaults:

- `ARGOS_ORG_ID=argos-local-org`
- `ARGOS_NEXUS_BASE_URL=http://localhost:18084`
- `ARGOS_NEXUS_API_KEY=argos-nexus-dev-key`
- `ARGOS_COMPANION_BASE_URL=http://localhost:18085`
- `ARGOS_COMPANION_API_KEY=argos-companion-dev-key`
- `ARGOS_PUBLIC_BASE_URL=http://localhost:13003`

Seed commands:

- `make seed-nexus-rules`
- `make seed-companion-assist`

If Nexus is down, Argos still stores the NDVI analysis. If Companion is down,
Nexus findings remain visible when available.

## Local Operation

Common commands:

- `make up`: build and start Docker services.
- `make down`: stop Docker services while preserving volumes.
- `make process-sample`: process `sample/` locally through the Python CLI.
- `make core-test`: run Go tests.
- `make processing-test`: run Python tests.
- `make ui-build`: build the UI.
- `make test`: run core, processing and UI checks.

Default local/Docker ports:

- UI Docker: `http://localhost:13003`
- Core API: `http://localhost:18090`
- Postgres: `15436`

Sample paths:

- Docker: `/data/sample`
- Local API process: `../sample`

## Possibly Obsolete / Pending Confirmation

- `docs/ARCHITECTURE.md` says Companion integration is kept behind future HTTP
  capabilities, but current code and README describe live Nexus/Companion
  integration paths.
- `docs/ROADMAP.md` lists read-only Companion capabilities under `Later`, but
  current code has assisted interpretation snapshots and seed scripts.

These items should be confirmed before deleting or rewriting those docs.

## Maintenance Log

- Created because `PROJECT_CONTEXT.md` did not exist. Content was derived from
  `README.md`, `Makefile`, `docker-compose.yml`, `.env.example`, `docs/`,
  migrations and the current repository layout.
- Updated after adding field CRUD and field-scoped dataset loading. The previous
  statement that formal field CRUD did not exist is now obsolete because
  `core/internal/fields` and migration `0005_fields` implement it.
