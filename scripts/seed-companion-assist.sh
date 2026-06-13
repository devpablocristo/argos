#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${ARGOS_COMPANION_BASE_URL:-http://localhost:18085}"
API_KEY="${ARGOS_COMPANION_API_KEY:-argos-companion-dev-key}"

curl -fsS -X POST "$BASE_URL/v1/assist-packs" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: $API_KEY" \
  --data-binary @- <<'JSON'
{
  "owner_system": "argos",
  "product_surface": "argos",
  "assist_type": "argos.analysis_explanation.v1",
  "name": "Argos analysis explanation",
  "description": "Explains Argos image-analysis facts and deterministic Nexus findings for human review.",
  "input_contract": "argos.analysis_assist_input.v1",
  "output_contract": "summary, simple_explanation, limitations, suggested_questions, next_steps",
  "prompt_template": "You are helping interpret an Argos multispectral drone analysis. Use the provided Argos facts and Nexus deterministic findings. Do not invent rules, do not prescribe agronomic treatment, and clearly separate observations from limitations. Return concise JSON with keys: summary, simple_explanation, limitations, suggested_questions, next_steps.\n\nInput JSON:\n{{input_json}}",
  "model_policy": {
    "max_tokens": 900
  },
  "enabled": true
}
JSON

echo "Seeded Argos-owned Companion assist pack."
