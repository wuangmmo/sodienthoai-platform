#!/usr/bin/env bash
set -euo pipefail
target="${1:?usage: scripts/rollback-code.sh <known-good-commit>}"
git cat-file -e "$target^{commit}"
git diff --quiet && git diff --cached --quiet || { echo "working_tree_not_clean"; exit 1; }
echo "Rollback changes application code only. Database migrations and persistent volumes are not automatically reversed."
git checkout --detach "$target"
COMPOSE_FILE="${COMPOSE_FILE:-docker-compose.production.yml}" scripts/deploy-production.sh
