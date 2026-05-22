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

When running with Docker, the sample dataset is mounted at `/data/sample`.
When running locally, use `../sample` from the API process.
