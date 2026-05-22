# Argos Roadmap

## Milestone 1: Local processing

- Python package `argos_processing`.
- DJI M3M capture group discovery.
- RGB + G/R/RE/NIR validation.
- NDVI calculation.
- PNG/TIFF/metadata export.
- Tests against `sample/`.

## Milestone 2: Core persistence

- Go API in `core`.
- PostgreSQL + PostGIS migrations.
- Dataset, capture, asset and analysis repositories.
- Adapter from Go to Python processing.

## Milestone 3: UI MVP

- Dataset list.
- Capture list.
- NDVI preview.
- Analysis outputs and downloads.
- Map only when georeferencing is trustworthy.

## Later

- Batch processing by flight/dataset.
- Orthomosaics and Cloud Optimized GeoTIFFs.
- NDRE/GNDVI and anomaly heuristics.
- Report generation.
- Model-assisted weed, disease and water-stress detection.
- Read-only Companion capabilities.

