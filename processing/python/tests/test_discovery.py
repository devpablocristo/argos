from pathlib import Path

from argos_processing.discovery import discover_capture_groups


def test_discovers_sample_capture_group() -> None:
    sample = Path(__file__).resolve().parents[3] / "sample"
    groups = discover_capture_groups(sample)

    assert len(groups) == 1
    group = groups[0]
    assert group.capture_key == "DJI_20230821133004_0001"
    assert group.validation.status == "valid"
    assert {asset.band for asset in group.assets} == {"rgb", "green", "red", "red_edge", "nir"}

