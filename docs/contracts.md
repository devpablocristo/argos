# Argos Internal Contracts

## `capture_group_manifest.v1`

Produced by `processing` and consumed by `core`.

```json
{
  "schema_version": "argos.eye.capture_group.v1",
  "vendor": "dji",
  "platform": "mavic_3_multispectral",
  "capture_key": "DJI_20230821133004_0001",
  "captured_at": "2023-08-21T13:30:04Z",
  "location": {
    "lat": 41.200977,
    "lon": -81.665370,
    "alt_m": 369.446,
    "crs": "EPSG:4326"
  },
  "assets": [],
  "validation": {
    "status": "valid",
    "warnings": []
  }
}
```

## `analysis_result.v1`

Produced by `processing` after NDVI export.

```json
{
  "schema_version": "argos.analysis_result.v1",
  "kind": "ndvi",
  "status": "completed",
  "outputs": [],
  "metrics": {
    "min": -1,
    "max": 1,
    "mean": 0.42
  }
}
```

