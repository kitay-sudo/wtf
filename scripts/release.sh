#!/usr/bin/env bash
# release.sh — bump SemVer, create git tag, push.
#
# Usage (from repo root or anywhere):
#   ./scripts/release.sh patch     v0.2.0 -> v0.2.1  (bug fixes)
#   ./scripts/release.sh minor     v0.2.0 -> v0.3.0  (new features)
#   ./scripts/release.sh major     v0.2.0 -> v1.0.0  (breaking changes)
#   ./scripts/release.sh v0.5.0    explicit version
#
# What happens after the tag is pushed:
#   release.yml workflow runs matrix build for 6 platforms
#   (linux/darwin/windows × amd64/arm64), uploads tar.gz / zip
#   archives + SHA256SUMS.txt to a fresh GitHub Release.

set -euo pipefail

if [ $# -lt 1 ]; then
  echo "Usage: $(basename "$0") <patch|minor|major|vX.Y.Z>" >&2
  exit 1
fi

# Always run from repo root.
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR/.."

# ---------- coloring ----------
if [ -t 1 ]; then
  C_OK=$'\e[33m'; C_WARN=$'\e[33m'; C_ERR=$'\e[31m'; C_DIM=$'\e[2m'
  C_BOLD=$'\e[1m'; C_CYAN=$'\e[36m'; C_RESET=$'\e[0m'
else
  C_OK=""; C_WARN=""; C_ERR=""; C_DIM=""; C_BOLD=""; C_CYAN=""; C_RESET=""
fi

ok()   { printf "  %s✓%s  %s\n" "$C_OK" "$C_RESET" "$*"; }
err()  { printf "  %s✗%s  %s\n" "$C_ERR" "$C_RESET" "$*" >&2; }
die()  { err "$*"; exit 1; }
step() { printf "  %s→%s  %s\n" "$C_CYAN" "$C_RESET" "$*"; }
info() { printf "  %sⓘ%s  %s\n" "$C_DIM" "$C_RESET" "$*"; }

# ---------- preflight ----------
BRANCH=$(git rev-parse --abbrev-ref HEAD)
if [ "$BRANCH" != "main" ]; then
  die "Release must be made from main, current: $BRANCH"
fi

if ! git diff --quiet HEAD; then
  die "Working tree is not clean. Commit or stash first."
fi
if ! git diff --cached --quiet; then
  die "Index has staged changes. Commit or reset first."
fi

step "git fetch --tags"
git fetch --tags --quiet

LOCAL=$(git rev-parse "@")
REMOTE=$(git rev-parse "@{u}" 2>/dev/null || echo "$LOCAL")
if [ "$LOCAL" != "$REMOTE" ]; then
  die "Local main does not match origin/main. Run: git pull --rebase OR git push"
fi

# ---------- last tag ----------
LAST_TAG=$(git tag --list "v*.*.*" --sort=-v:refname | head -1)
LAST_TAG=${LAST_TAG:-v0.0.0}
info "Last tag: $LAST_TAG"

# ---------- compute next ----------
BUMP="$1"

if [[ "$BUMP" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  NEXT="$BUMP"
else
  case "$BUMP" in
    patch|minor|major) ;;
    *) die "Unknown argument: $BUMP. Expected: patch | minor | major | vX.Y.Z" ;;
  esac

  VER=${LAST_TAG#v}
  IFS='.' read -r MAJ MIN PAT <<<"$VER"

  case "$BUMP" in
    major) MAJ=$((MAJ + 1)); MIN=0; PAT=0 ;;
    minor) MIN=$((MIN + 1)); PAT=0 ;;
    patch) PAT=$((PAT + 1)) ;;
  esac

  NEXT="v${MAJ}.${MIN}.${PAT}"
fi

if git rev-parse "$NEXT" >/dev/null 2>&1; then
  die "Tag $NEXT already exists."
fi

# ---------- show commits and ask confirmation ----------
echo
printf "  %sRelease:%s %s%s%s  (previous: %s)\n\n" "$C_BOLD" "$C_RESET" "$C_OK" "$NEXT" "$C_RESET" "$LAST_TAG"
echo "  Commits since last tag:"
git log --oneline --no-decorate "${LAST_TAG}..HEAD" | sed 's/^/    /'
echo

read -r -p "  Create tag $NEXT and push? [y/N]: " ANSWER
case "$ANSWER" in
  y|Y|yes|YES) ;;
  *) info "Cancelled."; exit 0 ;;
esac

# ---------- tag and push ----------
step "git tag -a $NEXT"
git tag -a "$NEXT" -m "$NEXT"

step "git push origin $NEXT"
git push origin "$NEXT"

echo
ok "Tag $NEXT pushed."
echo
echo "  Workflow release.yml has started. Track at:"
echo "    https://github.com/kitay-sudo/wtf/actions"
echo
echo "  In 2-3 minutes binaries appear at:"
echo "    https://github.com/kitay-sudo/wtf/releases/tag/$NEXT"
