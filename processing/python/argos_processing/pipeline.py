from __future__ import annotations

import json
from pathlib import Path

from .discovery import asset_by_band, discover_capture_groups
from .export import write_ndvi_png, write_ndvi_tiff
from .indices import calculate_ndvi, ndvi_metrics
from .schemas import AnalysisResult, OutputAsset, ProcessingResponse, SCHEMA_ANALYSIS_RESULT, to_plain


def process_capture_groups(input_path: Path, output_path: Path) -> ProcessingResponse:
    input_path = input_path.resolve()
    output_path = output_path.resolve()
    output_path.mkdir(parents=True, exist_ok=True)

    warnings: list[str] = []
    groups = discover_capture_groups(input_path)
    if not groups:
        warnings.append("no DJI M3M capture groups found")

    for group in groups:
        if group.validation.status != "valid":
            continue
        red = asset_by_band(group, "red")
        nir = asset_by_band(group, "nir")
        if red is None or nir is None:
            group.validation.errors.append("red and nir bands are required for NDVI")
            group.validation.status = "invalid"
            continue

        capture_output_dir = output_path / group.capture_key
        ndvi = calculate_ndvi(red.path, nir.path)
        png_path = capture_output_dir / "ndvi.png"
        tiff_path = capture_output_dir / "ndvi.tif"
        write_ndvi_png(ndvi, png_path)
        write_ndvi_tiff(ndvi, tiff_path)

        analysis = AnalysisResult(
            schema_version=SCHEMA_ANALYSIS_RESULT,
            kind="ndvi",
            status="completed",
            outputs=[
                OutputAsset(
                    kind="preview_png",
                    path=str(png_path),
                    content_type="image/png",
                    byte_size=png_path.stat().st_size,
                    metadata={"georeference_status": "metadata_only"},
                ),
                OutputAsset(
                    kind="analysis_tiff",
                    path=str(tiff_path),
                    content_type="image/tiff",
                    byte_size=tiff_path.stat().st_size,
                    metadata={"georeference_status": "metadata_only", "dtype": "float32"},
                ),
            ],
            metrics=ndvi_metrics(ndvi),
            warnings=["georeference_status=metadata_only"],
        )
        group.analysis = analysis
        metadata_path = capture_output_dir / "analysis_metadata.json"
        metadata_path.write_text(json.dumps(to_plain(analysis), indent=2, ensure_ascii=False) + "\n", encoding="utf-8")
        analysis.outputs.append(
            OutputAsset(
                kind="metadata_json",
                path=str(metadata_path),
                content_type="application/json",
                byte_size=metadata_path.stat().st_size,
                metadata={},
            )
        )

    status = "completed" if groups and any(g.validation.status == "valid" for g in groups) else "failed"
    return ProcessingResponse(
        schema_version="argos.processing_response.v1",
        status=status,
        input_path=str(input_path),
        output_path=str(output_path),
        captures=groups,
        warnings=warnings,
    )

