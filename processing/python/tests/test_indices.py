import numpy as np
import pytest

from argos_processing.indices import ndvi_metrics


def test_ndvi_metrics() -> None:
    metrics = ndvi_metrics(np.array([-1.0, 0.0, 1.0], dtype=np.float32))

    assert metrics["min"] == -1.0
    assert metrics["max"] == 1.0
    assert metrics["mean"] == 0.0
    assert metrics["valid_pixels"] == 3
    assert metrics["non_vegetation_pixels"] == 2
    assert metrics["high_vigor_pixels"] == 1
    assert metrics["non_vegetation_percent"] == pytest.approx(100 * 2 / 3)
    assert metrics["high_vigor_percent"] == pytest.approx(100 / 3)
