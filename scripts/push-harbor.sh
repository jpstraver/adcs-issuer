#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

if [[ -f "$ROOT_DIR/.env" ]]; then
  set -a
  # shellcheck disable=SC1091
  source "$ROOT_DIR/.env"
  set +a
fi

: "${HARBOR_USER:?HARBOR_USER is required (set in .env or environment)}"
: "${HARBOR_PASS:?HARBOR_PASS is required (set in .env or environment)}"

HARBOR_REGISTRY="${HARBOR_REGISTRY:-harbor.net.triviumpackaging.com}"
IMAGE="${HARBOR_IMAGE:-${HARBOR_REGISTRY}/pki/adcs-issuer}"
TAG="${1:-0.1.0}"
PUSH_LATEST="${PUSH_LATEST:-true}"

COMMIT="$(git -C "$ROOT_DIR" rev-parse --short HEAD 2>/dev/null || echo unknown)"
BUILD_TIME="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
PROJECT="${PROJECT:-github.com/djkormo/adcs-issuer}"
VERSION="${VERSION:-$TAG}"

printf '%s' "$HARBOR_PASS" | docker login "$HARBOR_REGISTRY" -u "$HARBOR_USER" --password-stdin

docker build \
  --build-arg VERSION="$VERSION" \
  --build-arg COMMIT="$COMMIT" \
  --build-arg BUILD_TIME="$BUILD_TIME" \
  --build-arg PROJECT="$PROJECT" \
  -t "$IMAGE:$TAG" \
  "$ROOT_DIR"

docker push "$IMAGE:$TAG"

if [[ "$PUSH_LATEST" == "true" ]]; then
  docker tag "$IMAGE:$TAG" "$IMAGE:latest"
  docker push "$IMAGE:latest"
  echo "Pushed: $IMAGE:$TAG and $IMAGE:latest"
else
  echo "Pushed: $IMAGE:$TAG"
fi
