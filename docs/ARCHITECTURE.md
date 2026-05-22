# Argos Architecture

Argos is a product repository, not a reusable platform module. It may consume
`platform` directly, while Companion integration is kept behind future HTTP
capabilities.

## Runtime split

- `core`: Go backend. Owns API, orchestration, job state, repositories and
  integration boundaries.
- `processing`: Python worker/CLI. Owns image reading, metadata extraction,
  validation, vegetation indices, raster exports and future ML inference.
- `ui`: React + TypeScript frontend. Owns dataset browsing and visualization.

The Python runtime is part of the Argos monorepo and is invoked by `core` using a
stable JSON contract. It is intentionally not an HTTP microservice in the MVP.
Docker can still isolate Python GIS dependencies in a dedicated image.

## Ecosystem service layout

`core/` follows the same service shape used by Nexus Governance:

```text
core/
  cmd/api/
  internal/
    catalog/
      handler.go
      handler/dto/
      repository.go
      usecases.go
      usecases/domain/
    analyses/
      handler.go
      handler/dto/
      repository.go
      usecases.go
      usecases/domain/
    processor/
    config/
  migrations/
  wire/
```

`catalog` and `analyses` are product domains. `processor` is a technical adapter
to the Python worker and should not accumulate domain behavior.

## Bounded contexts

- Catalog: datasets, flights, captures, band assets and metadata.
- Processing: processing jobs, validation, NDVI and future indices.
- Geodata: PostGIS, footprints, bounds and orthomosaics.
- Reports: analysis summaries and exports.
- Storage: original and derived files referenced by URI/path.
- Integrations: future Companion capabilities and event publication.

## Data flow

1. A dataset is registered in `core` with a local path or storage URI.
2. `core` creates a scan/process job and invokes `processing`.
3. `processing` discovers DJI capture groups and validates required assets.
4. `processing` calculates NDVI for valid capture groups.
5. `processing` writes PNG/TIFF/metadata outputs and returns JSON.
6. `core` stores normalized metadata and exposes assets through API endpoints.
7. `ui` renders capture metadata, preview imagery and analysis status.

## Georeferencing rule

Argos does not invent georeferencing. If a source raster exposes CRS/transform,
future Rasterio exports may preserve it. If only GPS metadata is available, the
result is marked as `metadata_only`.
