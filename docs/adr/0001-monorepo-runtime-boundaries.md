# ADR 0001: Monorepo With Go Core and Python Processing

## Status

Accepted.

## Decision

Argos starts as a monorepo with three first-class runtimes:

- `core` for Go orchestration and APIs.
- `processing` for Python CV/GIS/IA work.
- `ui` for React + TypeScript visualization.

The processing runtime is invoked as a CLI/worker with JSON contracts. It is not
an HTTP service in the MVP.

## Rationale

This keeps the first milestone small while preserving a strong boundary around
GDAL/Rasterio/OpenCV/NumPy dependencies. If processing later needs independent
scaling, the CLI contract can be promoted to a queued worker or internal HTTP
service without changing the product API.

## Consequences

- `core` remains the source of truth for job state.
- Python owns raster correctness and ML evolution.
- Docker images may be separated even though code remains in one repository.
- The MVP avoids Kubernetes, brokers and network hops between local modules.

