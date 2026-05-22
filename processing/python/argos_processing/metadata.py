from __future__ import annotations

import hashlib
import mimetypes
import re
from pathlib import Path
from typing import Any

from PIL import ExifTags, Image


XMP_TAG = 700


def sha256_file(path: Path) -> str:
    digest = hashlib.sha256()
    with path.open("rb") as fh:
        for chunk in iter(lambda: fh.read(1024 * 1024), b""):
            digest.update(chunk)
    return digest.hexdigest()


def mime_type(path: Path) -> str:
    guessed, _ = mimetypes.guess_type(path.name)
    if guessed:
        return guessed
    if path.suffix.lower() in {".tif", ".tiff"}:
        return "image/tiff"
    if path.suffix.lower() in {".jpg", ".jpeg"}:
        return "image/jpeg"
    return "application/octet-stream"


def image_info(path: Path) -> dict[str, Any]:
    with Image.open(path) as img:
        width, height = img.size
        bit_depth = _bit_depth(img)
        xmp = _extract_xmp(img)
        exif = _extract_exif(img)
    camera = parse_camera_xmp(xmp)
    gps = parse_gps(exif)
    return {
        "width": width,
        "height": height,
        "bit_depth": bit_depth,
        "camera": camera,
        "gps": gps,
        "exif": {
            "datetime": exif.get("DateTime") or exif.get("DateTimeOriginal"),
            "make": exif.get("Make"),
            "model": exif.get("Model"),
            "software": exif.get("Software"),
        },
    }


def parse_camera_xmp(xmp: str) -> dict[str, Any]:
    if not xmp:
        return {}
    out: dict[str, Any] = {}
    for key in [
        "BandName",
        "CentralWavelength",
        "WavelengthFWHM",
        "Irradiance",
        "IrradianceExposureTime",
        "IrradianceGain",
        "IsNormalized",
        "RigCameraIndex",
        "RigName",
        "SunSensor",
    ]:
        value = _find_xmp_value(xmp, f"Camera:{key}") or _find_xmp_value(xmp, f"drone-dji:{key}")
        if value is not None:
            out[key] = _coerce_value(value)
    return out


def parse_gps(exif: dict[str, Any]) -> dict[str, float | None]:
    gps = exif.get("GPSInfo")
    if not isinstance(gps, dict):
        return {"lat": None, "lon": None, "alt_m": None}
    lat = _gps_coord(gps.get("GPSLatitude"), gps.get("GPSLatitudeRef"))
    lon = _gps_coord(gps.get("GPSLongitude"), gps.get("GPSLongitudeRef"))
    alt = _ratio_to_float(gps.get("GPSAltitude"))
    return {"lat": lat, "lon": lon, "alt_m": alt}


def _bit_depth(img: Image.Image) -> int:
    if img.mode in {"I;16", "I;16L", "I;16B", "I"}:
        return 16
    if img.mode in {"F"}:
        return 32
    return 8


def _extract_xmp(img: Image.Image) -> str:
    raw = None
    if hasattr(img, "tag_v2"):
        raw = img.tag_v2.get(XMP_TAG)
    if raw is None:
        raw = img.info.get("XML:com.adobe.xmp") or img.info.get("xmp")
    if isinstance(raw, bytes):
        return raw.decode("utf-8", errors="ignore")
    if isinstance(raw, str):
        return raw
    return ""


def _extract_exif(img: Image.Image) -> dict[str, Any]:
    raw = img.getexif()
    out: dict[str, Any] = {}
    for tag_id, value in raw.items():
        tag = ExifTags.TAGS.get(tag_id, str(tag_id))
        if tag == "GPSInfo":
            out[tag] = _extract_gps_ifd(raw, value)
            continue
        out[tag] = value
    return out


def _extract_gps_ifd(raw: Image.Exif, value: Any) -> dict[str, Any]:
    gps_values = value if isinstance(value, dict) else {}
    if hasattr(raw, "get_ifd"):
        try:
            gps_values = raw.get_ifd(34853)
        except Exception:
            pass
    gps: dict[str, Any] = {}
    for gps_id, gps_value in gps_values.items():
        gps[ExifTags.GPSTAGS.get(gps_id, str(gps_id))] = gps_value
    return gps


def _find_xmp_value(xmp: str, key: str) -> str | None:
    patterns = [
        rf'{re.escape(key)}="([^"]+)"',
        rf"<{re.escape(key)}>(.*?)</{re.escape(key)}>",
    ]
    for pattern in patterns:
        match = re.search(pattern, xmp, flags=re.DOTALL)
        if match:
            return match.group(1).strip()
    return None


def _coerce_value(value: str) -> Any:
    value = value.strip()
    if value == "":
        return value
    if re.fullmatch(r"-?\d+", value):
        return int(value)
    if re.fullmatch(r"-?\d+\.\d+", value):
        return float(value)
    return value


def _ratio_to_float(value: Any) -> float | None:
    if value is None:
        return None
    try:
        return float(value)
    except (TypeError, ValueError):
        pass
    try:
        return float(value[0]) / float(value[1])
    except Exception:
        return None


def _gps_coord(value: Any, ref: Any) -> float | None:
    if not value:
        return None
    try:
        degrees = _ratio_to_float(value[0]) or 0.0
        minutes = _ratio_to_float(value[1]) or 0.0
        seconds = _ratio_to_float(value[2]) or 0.0
    except Exception:
        return None
    coord = degrees + minutes / 60.0 + seconds / 3600.0
    if str(ref).upper() in {"S", "W"}:
        coord *= -1
    return coord
