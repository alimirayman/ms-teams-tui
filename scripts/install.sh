#!/bin/sh
set -eu

repo=alimirayman/ms-teams-tui
install_dir=${MS_TEAMS_TUI_INSTALL_DIR:-"$HOME/.local/bin"}
requested_version=${MS_TEAMS_TUI_VERSION:-latest}

require_command() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "required command not found: $1" >&2
    exit 1
  fi
}

require_command curl

case $(uname -s) in
  Darwin) os=darwin ;;
  Linux) os=linux ;;
  *)
    echo "unsupported operating system: $(uname -s)" >&2
    exit 1
    ;;
esac

case $(uname -m) in
  x86_64|amd64) arch=amd64 ;;
  arm64|aarch64) arch=arm64 ;;
  *)
    echo "unsupported architecture: $(uname -m)" >&2
    exit 1
    ;;
esac

if [ "$requested_version" = "latest" ]; then
  release_url=$(curl -fsSL -o /dev/null -w '%{url_effective}' "https://github.com/$repo/releases/latest")
  tag=${release_url##*/}
else
  tag=$requested_version
  case "$tag" in
    v*) ;;
    *) tag=v$tag ;;
  esac
fi

version=${tag#v}
old_ifs=$IFS
IFS=.
set -- $version
IFS=$old_ifs
if [ "$#" -ne 3 ]; then
  echo "could not resolve a valid release version: $tag" >&2
  exit 1
fi
for part in "$@"; do
  case "$part" in
    0|[1-9]|[1-9][0-9]*) ;;
    *)
      echo "could not resolve a valid release version: $tag" >&2
      exit 1
      ;;
  esac
done

asset="ms-teams-tui-$tag-$os-$arch.tar.gz"
base_url="https://github.com/$repo/releases/download/$tag"
tmp_dir=$(mktemp -d "${TMPDIR:-/tmp}/ms-teams-tui-install.XXXXXX")
trap 'rm -rf "$tmp_dir"' EXIT HUP INT TERM

curl -fL "$base_url/$asset" -o "$tmp_dir/$asset"
curl -fL "$base_url/checksums.txt" -o "$tmp_dir/checksums.txt"

expected=$(awk -v asset="$asset" '$2 == asset { print $1 }' "$tmp_dir/checksums.txt")
if [ -z "$expected" ]; then
  echo "release checksum missing for $asset" >&2
  exit 1
fi

if command -v sha256sum >/dev/null 2>&1; then
  actual=$(sha256sum "$tmp_dir/$asset" | awk '{ print $1 }')
else
  require_command shasum
  actual=$(shasum -a 256 "$tmp_dir/$asset" | awk '{ print $1 }')
fi

if [ "$actual" != "$expected" ]; then
  echo "checksum verification failed for $asset" >&2
  exit 1
fi

tar -xzf "$tmp_dir/$asset" -C "$tmp_dir"
if [ ! -f "$tmp_dir/teams" ]; then
  echo "release archive does not contain the teams executable" >&2
  exit 1
fi

mkdir -p "$install_dir"
install -m 0755 "$tmp_dir/teams" "$install_dir/teams"

printf 'Installed ms-teams-tui %s as %s/teams\n' "$tag" "$install_dir"
case ":$PATH:" in
  *":$install_dir:"*) ;;
  *) printf 'Add %s to PATH before running teams.\n' "$install_dir" ;;
esac
