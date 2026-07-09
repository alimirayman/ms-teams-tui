#!/bin/sh
set -eu

kind=${1:-}
current=$(./scripts/check-version.sh)
old_ifs=$IFS
IFS=.
set -- $current
IFS=$old_ifs
major=$1
minor=$2
patch=$3

case "$kind" in
  patch)
    patch=$((patch + 1))
    ;;
  minor)
    minor=$((minor + 1))
    patch=0
    ;;
  major)
    major=$((major + 1))
    minor=0
    patch=0
    ;;
  *)
    echo "usage: $0 patch|minor|major" >&2
    exit 2
    ;;
esac

next=$major.$minor.$patch
printf '%s\n' "$next" > VERSION
printf 'VERSION: %s -> %s\n' "$current" "$next"
