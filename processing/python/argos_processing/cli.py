from __future__ import annotations

import argparse
import json
import sys
from pathlib import Path

from .pipeline import process_capture_groups
from .schemas import to_plain


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser(prog="argos-processing")
    subcommands = parser.add_subparsers(dest="command", required=True)

    process = subcommands.add_parser("process-capture-group", help="Discover, validate and process DJI capture groups")
    process.add_argument("--input", required=True, help="Input directory containing DJI capture files")
    process.add_argument("--output", required=True, help="Output directory for derived assets")
    process.add_argument("--json-output", help="Optional path where the response JSON should be written")

    args = parser.parse_args(argv)
    if args.command == "process-capture-group":
        return _process(args)
    parser.error("unknown command")
    return 2


def _process(args: argparse.Namespace) -> int:
    try:
        response = process_capture_groups(Path(args.input), Path(args.output))
        body = json.dumps(to_plain(response), indent=2, ensure_ascii=False) + "\n"
        if args.json_output:
            Path(args.json_output).parent.mkdir(parents=True, exist_ok=True)
            Path(args.json_output).write_text(body, encoding="utf-8")
        sys.stdout.write(body)
        return 0 if response.status == "completed" else 1
    except Exception as exc:
        error = {
            "schema_version": "argos.processing_error.v1",
            "status": "failed",
            "error": str(exc),
        }
        sys.stderr.write(json.dumps(error, ensure_ascii=False) + "\n")
        return 1


if __name__ == "__main__":
    raise SystemExit(main())

