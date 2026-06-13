# Argos Agent Guide

This file is for Codex and other coding agents working in this repository.
Keep it factual and incremental. Do not replace it wholesale when updating it.

## Mandatory Maintenance Rule

If `AGENTS.md` or `PROJECT_CONTEXT.md` already exist, treat them as primary
project context before editing.

Required process:

1. Read `AGENTS.md` and `PROJECT_CONTEXT.md` completely if they exist.
2. Preserve useful rules, commands, warnings, conventions and prior decisions.
3. Improve structure and clarity without deleting valuable meaning.
4. Merge duplicated or disordered information carefully.
5. If something appears obsolete, move it to `Possibly obsolete / pending confirmation`
   or leave a note explaining the doubt.
6. Only remove content when it is clearly duplicated, contradictory or incorrect
   according to repo evidence.
7. If a section is removed or replaced, explain what changed and why at the end.
8. Do not invent rules that are not backed by code, docs, configs or an explicit
   user decision.

## Product Summary

Argos is an agricultural vision system for multispectral drone datasets. The
current target is DJI Mavic 3 Multispectral capture groups with RGB plus
green, red, red-edge and NIR bands.

The repository is a pragmatic product monorepo:

- `core`: Go API, orchestration, persistence, migrations and integration
  boundaries.
- `processing`: Python package and CLI worker for image reading, metadata
  extraction, validation, NDVI and raster exports.
- `ui`: React + TypeScript frontend for dataset browsing and visualization.
- `docs`: architecture notes, contracts, ADRs and roadmap.
- `sample`: local DJI sample capture group.

`core` follows the ecosystem service layout used by projects such as Nexus:
domain folders live under `internal/`, with `handler`, `handler/dto`,
`repository`, `usecases` and `usecases/domain`, plus service-level `wire/` and
`migrations/`.

## Common Commands

- `make up`: create the `axis-local` Docker network if needed, then build and
  start Argos containers.
- `make down`: stop containers. Docker volumes are kept.
- `make core-test`: run Go tests in `core`.
- `make processing-test`: install the Python processing package in its local
  venv and run pytest.
- `make ui-build`: build the React UI.
- `make test`: run core, processing and UI checks.
- `make process-sample`: run the Python processing CLI against `sample/`.
- `make seed-nexus-rules`: seed Argos-owned finding rules into Nexus.
- `make seed-companion-assist`: seed Argos-owned assist packs into Companion.
- `make run-core-dev`: run the Go core with Air hot reload.
- `make run-ui`: run the UI dev server.

## Runtime Defaults

- Docker UI: `http://localhost:13003`.
- Docker core API: `http://localhost:18090`.
- Docker Postgres/PostGIS host port: `15436`.
- Docker sample path: `/data/sample`.
- Local sample path from the API process: `../sample`.
- Docker storage/output path: `/data/outputs`.
- Local storage/output default: `../var/outputs`.

Docker starts PostgreSQL/PostGIS. `core` applies migrations on startup. The
Docker backend runs with Air hot reload.

## Axis Integration

Argos can publish analysis facts to Nexus and request assisted interpretation
from Companion. Locally, start Axis before Argos:

```bash
cd ../axis && make up
cd ../argos && make up
make seed-nexus-rules
make seed-companion-assist
```

Default local integration values:

- `ARGOS_ORG_ID=argos-local-org`
- `ARGOS_NEXUS_BASE_URL=http://localhost:18084`
- `ARGOS_NEXUS_API_KEY=argos-nexus-dev-key`
- `ARGOS_COMPANION_BASE_URL=http://localhost:18085`
- `ARGOS_COMPANION_API_KEY=argos-companion-dev-key`
- `ARGOS_PUBLIC_BASE_URL=http://localhost:13003`

Nexus produces deterministic findings. Companion produces assisted AI
interpretation. If Nexus or Companion is unavailable, Argos still stores the
NDVI analysis and surfaces degraded sync state in the UI.

## Domain Rules

- A dataset is a raw-file ingestion unit. It is not automatically a field, lot,
  campaign, flight or capture.
- Users should start by creating or choosing a field, then registering or
  uploading datasets under that field.
- Argos now has formal field CRUD in v1. Lot, campaign and flight CRUD remain
  future work.
- The dataset list remains a technical history of datasets processed by this
  Argos installation. Field-scoped dataset history should only show datasets
  with a real `field_id`.
- Dataset semantic classification is owned by Argos.
- Classification fields for field, lot, campaign and flight may be null. Argos
  must not invent these associations.
- Captures are coherent image/band groups derived from datasets.
- Analyses are generated results such as NDVI, RGB preview assets, findings and
  assisted interpretation snapshots.
- Argos does not invent georeferencing. If only GPS metadata is available, the
  result is metadata-only context.

## Data and Lifecycle Notes

- Dataset history follows CRUDAR:
  - Create: create/select a field, then scan or upload a source folder into a
    dataset under that field.
  - Read: reopen saved datasets, captures and analyses.
  - Update: edit dataset name or source URI.
  - Archive/Restore: hide or recover datasets without losing analysis data.
  - Delete: after archive, physically removes dataset rows and Argos generated
    outputs; it does not delete original source images.
- `make down` preserves Docker volumes, so Postgres data survives the next
  `make up`.
- `dataset_classifications` stores semantic classification.
- `dataset_events` stores append-only dataset history events.
- `fields` stores productive fields/establishments used to group datasets.
- Dataset rows may reference `fields.id` through nullable `field_id`.

## Possibly Obsolete / Pending Confirmation

- `docs/ARCHITECTURE.md` says Companion integration is future-facing, while the
  current code already has Nexus and Companion integration paths.
- `docs/ROADMAP.md` lists read-only Companion capabilities under `Later`, while
  the current code has assisted interpretation snapshots.

These notes should not be deleted without checking the current code and docs.

## Maintenance Log

- Created because the file did not exist. Content was derived from `README.md`,
  `Makefile`, `docker-compose.yml`, `.env.example`, `docs/`, migrations and the
  current repository layout.
- Updated after adding field CRUD and field-first dataset loading. The old note
  that formal field CRUD did not exist was replaced because `core/internal/fields`
  and migration `0005_fields` now implement it.
