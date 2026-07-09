#!/bin/sh
set -eu

version=$(tr -d '[:space:]' < VERSION)

case "$version" in
  ''|*[!0-9.]*|.*|*.)
    echo "VERSION must contain semantic version X.Y.Z" >&2
    exit 1
    ;;
esac

old_ifs=$IFS
IFS=.
set -- $version
IFS=$old_ifs

if [ "$#" -ne 3 ] || [ -z "$1" ] || [ -z "$2" ] || [ -z "$3" ]; then
  echo "VERSION must contain semantic version X.Y.Z" >&2
  exit 1
fi

for part in "$@"; do
  case "$part" in
    0|[1-9]|[1-9][0-9]*) ;;
    *)
      echo "VERSION components must be non-negative integers without leading zeroes" >&2
      exit 1
      ;;
  esac
done

printf '%s\n' "$version"
