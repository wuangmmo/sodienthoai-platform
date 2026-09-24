#!/usr/bin/env bash
set -euo pipefail
compose="${COMPOSE_FILE:-docker-compose.production.yml}"
project="${COMPOSE_PROJECT_NAME:-sodienthoai-production}"
api_port="${PRODUCTION_API_PORT:-19080}"
test -f "$compose"
git diff --quiet && git diff --cached --quiet || { echo "working_tree_not_clean"; exit 1; }
sha="$(git rev-parse HEAD)"
echo "deploy_commit=$sha"
echo "deploy_project=$project"
echo "deploy_api_port=$api_port"
docker compose -p "$project" -f "$compose" config --quiet
docker compose -p "$project" -f "$compose" build
docker compose -p "$project" -f "$compose" up -d --remove-orphans
for i in $(seq 1 30); do
  body="$(curl -fsS "http://127.0.0.1:$api_port/readyz" 2>/dev/null || true)"
  echo "$body" | grep -q '"status":"ready"' && { echo "deploy_ready=$sha"; exit 0; }
  sleep 2
done
docker compose -p "$project" -f "$compose" ps
echo "deploy_readiness_failed=$sha"
exit 1
