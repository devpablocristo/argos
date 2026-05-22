from __future__ import annotations

from pathlib import Path

import numpy as np
from PIL import Image
import tifffile


def write_ndvi_png(ndvi: np.ndarray, path: Path) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    rgb = colorize_ndvi(ndvi)
    Image.fromarray(rgb, mode="RGB").save(path)


def write_ndvi_tiff(ndvi: np.ndarray, path: Path) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    tifffile.imwrite(path, ndvi.astype(np.float32), metadata={"kind": "ndvi", "georeference_status": "metadata_only"})


def colorize_ndvi(ndvi: np.ndarray) -> np.ndarray:
    values = np.nan_to_num(ndvi, nan=-1.0, posinf=1.0, neginf=-1.0)
    normalized = ((values + 1.0) / 2.0).clip(0.0, 1.0)

    stops = np.array(
        [
            [72, 38, 26],
            [189, 102, 47],
            [238, 218, 115],
            [106, 168, 79],
            [28, 97, 45],
        ],
        dtype=np.float32,
    )
    scaled = normalized * (len(stops) - 1)
    low = np.floor(scaled).astype(np.int32)
    high = np.clip(low + 1, 0, len(stops) - 1)
    frac = (scaled - low)[..., None]
    rgb = stops[low] * (1.0 - frac) + stops[high] * frac
    return rgb.astype(np.uint8)

