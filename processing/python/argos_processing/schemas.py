from __future__ import annotations

from dataclasses import asdict, dataclass, field
from typing import Any


SCHEMA_CAPTURE_GROUP = "argos.eye.capture_group.v1"
SCHEMA_ANALYSIS_RESULT = "argos.analysis_result.v1"


@dataclass
class Location:
    lat: float | None = None
    lon: float | None = None
    alt_m: float | None = None
    crs: str = "EPSG:4326"


@dataclass
class BandAsset:
    band: str
    role: str
    path: str
    checksum_sha256: str
    mime_type: str
    width: int
    height: int
    bit_depth: int
    wavelength_nm: int | None = None
    fwhm_nm: int | None = None
    source_metadata: dict[str, Any] = field(default_factory=dict)


@dataclass
class Validation:
    status: str
    warnings: list[str] = field(default_factory=list)
    errors: list[str] = field(default_factory=list)


@dataclass
class OutputAsset:
    kind: str
    path: str
    content_type: str
    byte_size: int
    metadata: dict[str, Any] = field(default_factory=dict)


@dataclass
class AnalysisResult:
    schema_version: str
    kind: str
    status: str
    outputs: list[OutputAsset]
    metrics: dict[str, Any]
    warnings: list[str] = field(default_factory=list)


@dataclass
class CaptureGroupManifest:
    schema_version: str
    vendor: str
    platform: str
    capture_key: str
    captured_at: str | None
    location: Location
    assets: list[BandAsset]
    validation: Validation
    analysis: AnalysisResult | None = None


@dataclass
class ProcessingResponse:
    schema_version: str
    status: str
    input_path: str
    output_path: str
    captures: list[CaptureGroupManifest]
    warnings: list[str] = field(default_factory=list)


def to_plain(value: Any) -> Any:
    return asdict(value)

