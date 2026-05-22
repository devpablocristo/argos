from __future__ import annotations

from pathlib import Path

import numpy as np
from PIL import Image


def read_band(path: str | Path) -> np.ndarray:
    with Image.open(path) as img:
        return np.asarray(img, dtype=np.float32)


def calculate_ndvi(red_path: str | Path, nir_path: str | Path) -> np.ndarray:
    red = read_band(red_path)
    nir = read_band(nir_path)
    if red.shape != nir.shape:
        raise ValueError(f"red/nir shape mismatch: {red.shape} != {nir.shape}")

    denominator = nir + red
    ndvi = np.divide(nir - red, denominator, out=np.zeros_like(nir, dtype=np.float32), where=np.abs(denominator) > 1e-6)
    return np.clip(ndvi, -1.0, 1.0).astype(np.float32)


def ndvi_metrics(ndvi: np.ndarray) -> dict[str, float]:
    finite = ndvi[np.isfinite(ndvi)]
    if finite.size == 0:
        return {
            "min": 0.0,
            "max": 0.0,
            "mean": 0.0,
            "std": 0.0,
            "valid_pixels": 0,
            "non_vegetation_percent": 0.0,
            "low_vigor_percent": 0.0,
            "medium_vigor_percent": 0.0,
            "high_vigor_percent": 0.0,
            "non_vegetation_pixels": 0,
            "low_vigor_pixels": 0,
            "medium_vigor_pixels": 0,
            "high_vigor_pixels": 0,
        }

    non_vegetation = finite < 0.2
    low_vigor = (finite >= 0.2) & (finite < 0.4)
    medium_vigor = (finite >= 0.4) & (finite < 0.6)
    high_vigor = finite >= 0.6
    total = finite.size

    metrics = {
        "min": float(np.min(finite)),
        "max": float(np.max(finite)),
        "mean": float(np.mean(finite)),
        "std": float(np.std(finite)),
        "valid_pixels": int(finite.size),
        "non_vegetation_pixels": int(np.count_nonzero(non_vegetation)),
        "low_vigor_pixels": int(np.count_nonzero(low_vigor)),
        "medium_vigor_pixels": int(np.count_nonzero(medium_vigor)),
        "high_vigor_pixels": int(np.count_nonzero(high_vigor)),
    }
    metrics["non_vegetation_percent"] = _percent(metrics["non_vegetation_pixels"], total)
    metrics["low_vigor_percent"] = _percent(metrics["low_vigor_pixels"], total)
    metrics["medium_vigor_percent"] = _percent(metrics["medium_vigor_pixels"], total)
    metrics["high_vigor_percent"] = _percent(metrics["high_vigor_pixels"], total)
    return metrics


def _percent(count: int, total: int) -> float:
    if total <= 0:
        return 0.0
    return float(count / total * 100.0)
