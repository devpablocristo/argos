# Argos

Argos is an agricultural vision system for multispectral drone datasets. The
initial target is DJI Mavic 3 Multispectral capture groups containing RGB plus
G/R/RE/NIR bands.

The repository is intentionally a pragmatic monorepo:

```text
argos/
  core/          # Go API and orchestration
  processing/    # Python CV/GIS/IA worker/CLI
  ui/            # React + TypeScript frontend
  docs/          # architecture decisions and roadmap
  sample/        # local DJI sample capture group
```

`core/` follows the ecosystem service layout used by projects such as Nexus:
domain folders under `internal/`, each with `handler`, `usecases`, `repository`
and `usecases/domain`, plus service-level `wire/` and `migrations/`.

## Milestone 1

- Discover DJI M3M capture groups.
- Validate RGB + G/R/RE/NIR integrity.
- Calculate NDVI: `(NIR - RED) / (NIR + RED)`.
- Export preview PNG, analysis TIFF and metadata JSON.
- Expose the basic workflow through `core` and `ui`.

## Local processing

```bash
make process-sample
```

## Local API

```bash
make processing-install
cd core
go run ./cmd/api
```

Default API URL: `http://localhost:18090`.

`core` automatically uses `processing/python/.venv/bin/python` when it exists.

For Go hot reload during development:

```bash
make air-install
make run-core-dev
```

## Local UI

```bash
cd ui
npm install
npm run dev
```

Default UI URL: `http://localhost:5173`.

## Docker

```bash
make up
```

Default Docker UI URL: `http://localhost:13003`.
The Docker backend runs with Air hot reload.
Docker also starts PostgreSQL/PostGIS and `core` applies migrations on startup.

When running with Docker, the sample dataset is mounted at `/data/sample`.
When running locally, use `../sample` from the API process.

`make down` stops containers but keeps Docker volumes, so datasets, captures and
analyses stored in Postgres survive the next `make up`.

## Product workflow

Argos starts from the agronomic context before loading files:

1. Create or choose a field.
2. Load images or register a source path as a dataset under that field.
3. Scan the dataset to detect captures.
4. Generate NDVI and review outputs, findings and assisted interpretation.

A dataset is still the raw-file ingestion unit. It is not automatically a field,
lot, campaign, flight or capture.

## Axis integration

Argos can publish analysis facts to Nexus and request assisted interpretation
from Companion. Locally, start Axis first, then Argos:

```bash
cd ../axis && make up
cd ../argos && make up
make seed-nexus-rules
make seed-companion-assist
```

Defaults:

- `ARGOS_ORG_ID=argos-local-org`
- `ARGOS_NEXUS_BASE_URL=http://localhost:18084`
- `ARGOS_NEXUS_API_KEY=argos-nexus-dev-key`
- `ARGOS_COMPANION_BASE_URL=http://localhost:18085`
- `ARGOS_COMPANION_API_KEY=argos-companion-dev-key`

If Nexus or Companion is down, Argos still stores the NDVI analysis and shows
the degraded sync status in the UI.

The dataset history follows a CRUDAR model:

- Create: create/select a field, then scan a source folder into a dataset.
- Read: reopen saved datasets, captures and analyses.
- Update: edit dataset name or source URI.
- Archive/Restore: hide or recover datasets without losing analysis data.
- Delete: after archive, physically removes the dataset rows and Argos generated
  outputs; it does not delete the original source images.
