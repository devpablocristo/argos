# Argos Core

`core/` is the Go backend service for Argos. It follows the same internal shape
used by ecosystem services such as Nexus Governance:

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

`processor/` is an adapter to the Python worker in `../processing`; it is not a
domain module.

## Run

```bash
go run ./cmd/api
```

Default URL: `http://localhost:18090`.

