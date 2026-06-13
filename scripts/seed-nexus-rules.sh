#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${ARGOS_NEXUS_BASE_URL:-http://localhost:18084}"
API_KEY="${ARGOS_NEXUS_API_KEY:-argos-nexus-dev-key}"

post_rule() {
  curl -fsS -X POST "$BASE_URL/v1/finding-rules" \
    -H "Content-Type: application/json" \
    -H "X-API-Key: $API_KEY" \
    --data-binary @-
}

post_rule <<'JSON'
{
  "owner_system": "argos",
  "source_system": "argos",
  "fact_type": "argos.analysis_facts.v1",
  "code": "argos.ndvi.low_vigor_high",
  "name": "High low-vigor area",
  "description": "Flags analyses where a large share of valid pixels falls in the low NDVI vigor band.",
  "expression": "double(facts[\"low_vigor_percent\"]) >= 35.0",
  "severity": "warning",
  "title": "Large low-vigor area",
  "message": "A significant portion of the capture is in the low-vigor NDVI range.",
  "recommendation": "Review the affected area against field context, irrigation history, soil data and recent operations.",
  "mode": "enforced",
  "enabled": true,
  "priority": 10
}
JSON

post_rule <<'JSON'
{
  "owner_system": "argos",
  "source_system": "argos",
  "fact_type": "argos.analysis_facts.v1",
  "code": "argos.ndvi.non_vegetation_high",
  "name": "High non-vegetation area",
  "description": "Flags analyses where non-vegetation or exposed-soil NDVI is elevated.",
  "expression": "double(facts[\"non_vegetation_percent\"]) >= 20.0",
  "severity": "info",
  "title": "Elevated non-vegetation area",
  "message": "The analysis contains a notable amount of pixels outside the vegetation NDVI range.",
  "recommendation": "Check whether paths, water, shadows, bare soil or image boundaries are expected for this capture.",
  "mode": "enforced",
  "enabled": true,
  "priority": 20
}
JSON

post_rule <<'JSON'
{
  "owner_system": "argos",
  "source_system": "argos",
  "fact_type": "argos.analysis_facts.v1",
  "code": "argos.ndvi.high_vigor_low",
  "name": "Low high-vigor area",
  "description": "Flags analyses where high NDVI vigor is scarce.",
  "expression": "double(facts[\"high_vigor_percent\"]) <= 15.0",
  "severity": "warning",
  "title": "Limited high-vigor area",
  "message": "Only a small portion of the capture is in the high-vigor NDVI range.",
  "recommendation": "Compare this result with crop stage, expected canopy coverage and prior captures before deciding an action.",
  "mode": "enforced",
  "enabled": true,
  "priority": 30
}
JSON

echo "Seeded Argos-owned Nexus finding rules."
