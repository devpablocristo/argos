from __future__ import annotations

import re
from pathlib import Path

from .metadata import image_info, mime_type, sha256_file
from .schemas import BandAsset, CaptureGroupManifest, Location, SCHEMA_CAPTURE_GROUP, Validation


DJI_RE = re.compile(r"^(?P<key>DJI_\d{14}_\d{4})(?P<suffix>_D|_MS_G|_MS_R|_MS_RE|_MS_NIR)\.(?P<ext>JPG|JPEG|TIF|TIFF)$", re.IGNORECASE)

BAND_BY_SUFFIX = {
    "_D": "rgb",
    "_MS_G": "green",
    "_MS_R": "red",
    "_MS_RE": "red_edge",
    "_MS_NIR": "nir",
}

ROLE_BY_BAND = {
    "rgb": "rgb",
    "green": "multispectral_band",
    "red": "multispectral_band",
    "red_edge": "multispectral_band",
    "nir": "multispectral_band",
}

WAVELENGTH_BY_BAND = {
    "green": 560,
    "red": 650,
    "red_edge": 730,
    "nir": 860,
}

REQUIRED_BANDS = {"rgb", "green", "red", "red_edge", "nir"}
MS_BANDS = {"green", "red", "red_edge", "nir"}


def discover_capture_groups(input_path: Path) -> list[CaptureGroupManifest]:
    files_by_key: dict[str, dict[str, Path]] = {}
    for path in sorted(input_path.iterdir()):
        if not path.is_file():
            continue
        match = DJI_RE.match(path.name)
        if not match:
            continue
        band = BAND_BY_SUFFIX[match.group("suffix").upper()]
        files_by_key.setdefault(match.group("key"), {})[band] = path

    groups: list[CaptureGroupManifest] = []
    for capture_key, files in sorted(files_by_key.items()):
        groups.append(_build_group(capture_key, files))
    return groups


def _build_group(capture_key: str, files: dict[str, Path]) -> CaptureGroupManifest:
    warnings: list[str] = []
    errors: list[str] = []
    assets: list[BandAsset] = []
    captured_at: str | None = None
    location = Location()

    missing = sorted(REQUIRED_BANDS - set(files))
    if missing:
        errors.append("missing required bands: " + ", ".join(missing))

    ms_shape: tuple[int, int, int] | None = None
    for band in ["rgb", "green", "red", "red_edge", "nir"]:
        path = files.get(band)
        if path is None:
            continue
        info = image_info(path)
        camera = info.get("camera") or {}
        gps = info.get("gps") or {}
        exif = info.get("exif") or {}

        if captured_at is None:
            captured_at = _normalize_dji_time(exif.get("datetime"))
        if location.lat is None:
            location = Location(lat=gps.get("lat"), lon=gps.get("lon"), alt_m=gps.get("alt_m"))

        width = int(info["width"])
        height = int(info["height"])
        bit_depth = int(info["bit_depth"])
        if band in MS_BANDS:
            shape = (width, height, bit_depth)
            if ms_shape is None:
                ms_shape = shape
            elif shape != ms_shape:
                errors.append(f"multispectral shape mismatch for {band}: {shape} != {ms_shape}")

        wavelength = _int_or_none(camera.get("CentralWavelength")) or WAVELENGTH_BY_BAND.get(band)
        fwhm = _int_or_none(camera.get("WavelengthFWHM"))
        band_name = str(camera.get("BandName") or "").strip().lower()
        if band != "rgb" and band_name and _normalize_band_name(band_name) != band:
            warnings.append(f"metadata band name mismatch for {band}: {band_name}")

        assets.append(
            BandAsset(
                band=band,
                role=ROLE_BY_BAND[band],
                path=str(path.resolve()),
                checksum_sha256=sha256_file(path),
                mime_type=mime_type(path),
                width=width,
                height=height,
                bit_depth=bit_depth,
                wavelength_nm=wavelength,
                fwhm_nm=fwhm,
                source_metadata={"camera": camera, "gps": gps, "exif": exif},
            )
        )

    if not errors and ms_shape is None:
        errors.append("no multispectral bands found")
    status = "valid" if not errors else "invalid"
    return CaptureGroupManifest(
        schema_version=SCHEMA_CAPTURE_GROUP,
        vendor="dji",
        platform="mavic_3_multispectral",
        capture_key=capture_key,
        captured_at=captured_at,
        location=location,
        assets=assets,
        validation=Validation(status=status, warnings=warnings, errors=errors),
    )


def asset_by_band(group: CaptureGroupManifest, band: str) -> BandAsset | None:
    for asset in group.assets:
        if asset.band == band:
            return asset
    return None


def _normalize_dji_time(value: object) -> str | None:
    if not value:
        return None
    raw = str(value).strip()
    match = re.fullmatch(r"(\d{4}):(\d{2}):(\d{2}) (\d{2}):(\d{2}):(\d{2})", raw)
    if not match:
        return raw
    year, month, day, hour, minute, second = match.groups()
    return f"{year}-{month}-{day}T{hour}:{minute}:{second}Z"


def _normalize_band_name(value: str) -> str:
    value = value.replace(" ", "").replace("-", "").replace("_", "").lower()
    if value == "rededge":
        return "red_edge"
    if value == "nir":
        return "nir"
    if value == "red":
        return "red"
    if value == "green":
        return "green"
    return value


def _int_or_none(value: object) -> int | None:
    if value is None or value == "":
        return None
    try:
        return int(float(str(value)))
    except ValueError:
        return None

