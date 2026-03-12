#!/bin/sh
# Panen installer — downloads and installs the latest release.
# Usage: curl -fsSL https://raw.githubusercontent.com/lugassawan/panen/main/scripts/install.sh | sh
#
# Environment variables:
#   PANEN_VERSION  — specific version to install (e.g. "v0.2.0"); defaults to latest

set -eu

REPO="lugassawan/panen"
GITHUB_API="https://api.github.com"
GITHUB_RELEASES="https://github.com/${REPO}/releases/download"

# --- Helpers ---

log() {
  printf '%s\n' "$@"
}

fail() {
  printf 'Error: %s\n' "$1" >&2
  exit 1
}

need_cmd() {
  if ! command -v "$1" > /dev/null 2>&1; then
    fail "required command not found: $1"
  fi
}

# --- Detect platform ---

detect_platform() {
  OS=$(uname -s)
  ARCH=$(uname -m)

  case "$OS" in
    Darwin)
      PLATFORM="darwin"
      ARCHIVE="panen-darwin-universal.zip"
      ;;
    Linux)
      PLATFORM="linux"
      case "$ARCH" in
        x86_64|amd64)
          ARCHIVE="panen-linux-amd64.tar.gz"
          ;;
        *)
          fail "unsupported Linux architecture: $ARCH"
          ;;
      esac
      ;;
    *)
      fail "unsupported operating system: $OS (use macOS or Linux)"
      ;;
  esac

  log "Detected platform: ${PLATFORM} (${ARCH})"
}

# --- Resolve version ---

resolve_version() {
  if [ -n "${PANEN_VERSION:-}" ]; then
    VERSION="$PANEN_VERSION"
    # Ensure version starts with "v"
    case "$VERSION" in
      v*) ;;
      *)  VERSION="v${VERSION}" ;;
    esac
  else
    need_cmd curl
    log "Fetching latest release..."
    VERSION=$(curl -fsSL "${GITHUB_API}/repos/${REPO}/releases/latest" | grep '"tag_name"' | sed -E 's/.*"tag_name": *"([^"]+)".*/\1/')
    if [ -z "$VERSION" ]; then
      fail "could not determine latest version"
    fi
  fi

  log "Installing Panen ${VERSION}"
}

# --- Download and verify ---

download_and_verify() {
  WORK_DIR=$(mktemp -d)
  trap 'rm -rf "$WORK_DIR"' EXIT

  ARCHIVE_URL="${GITHUB_RELEASES}/${VERSION}/${ARCHIVE}"
  CHECKSUMS_URL="${GITHUB_RELEASES}/${VERSION}/SHA256SUMS.txt"

  log "Downloading ${ARCHIVE}..."
  curl -fsSL -o "${WORK_DIR}/${ARCHIVE}" "$ARCHIVE_URL"

  log "Downloading checksums..."
  curl -fsSL -o "${WORK_DIR}/SHA256SUMS.txt" "$CHECKSUMS_URL"

  log "Verifying checksum..."
  cd "$WORK_DIR"
  if command -v sha256sum > /dev/null 2>&1; then
    if ! grep "$ARCHIVE" SHA256SUMS.txt | sha256sum -c - > /dev/null 2>&1; then
      fail "checksum verification failed for ${ARCHIVE}"
    fi
  elif command -v shasum > /dev/null 2>&1; then
    if ! grep "$ARCHIVE" SHA256SUMS.txt | shasum -a 256 -c - > /dev/null 2>&1; then
      fail "checksum verification failed for ${ARCHIVE}"
    fi
  else
    fail "no SHA256 checksum tool found (need sha256sum or shasum)"
  fi
  log "Checksum verified"
}

# --- Install ---

install_darwin() {
  INSTALL_DIR="/Applications"

  if [ ! -w "$INSTALL_DIR" ]; then
    fail "/Applications is not writable. Run with sudo or ensure you are an admin user."
  fi

  log "Extracting to ${INSTALL_DIR}/Panen.app..."
  rm -rf "${INSTALL_DIR}/panen.app"
  rm -rf "${INSTALL_DIR}/Panen.app"
  unzip -q "${WORK_DIR}/${ARCHIVE}" -d "$INSTALL_DIR"

  # Clean up old ~/Applications installs (migration for existing users)
  rm -rf "${HOME}/Applications/panen.app"
  rm -rf "${HOME}/Applications/Panen.app"

  log ""
  log "Panen has been installed to ${INSTALL_DIR}/Panen.app"
  log "You can launch it from /Applications or Spotlight."
}

install_linux() {
  BIN_DIR="${HOME}/.local/bin"
  DESKTOP_DIR="${HOME}/.local/share/applications"
  ICON_DIR="${HOME}/.local/share/icons"

  mkdir -p "$BIN_DIR" "$DESKTOP_DIR" "$ICON_DIR"

  log "Extracting..."
  tar -xzf "${WORK_DIR}/${ARCHIVE}" -C "$WORK_DIR"

  cp "${WORK_DIR}/panen/panen" "${BIN_DIR}/panen"
  chmod +x "${BIN_DIR}/panen"

  cp "${WORK_DIR}/panen/panen.desktop" "${DESKTOP_DIR}/panen.desktop"
  cp "${WORK_DIR}/panen/panen.png" "${ICON_DIR}/panen.png"

  log ""
  log "Panen has been installed:"
  log "  Binary:  ${BIN_DIR}/panen"
  log "  Desktop: ${DESKTOP_DIR}/panen.desktop"
  log "  Icon:    ${ICON_DIR}/panen.png"

  case ":${PATH}:" in
    *":${BIN_DIR}:"*) ;;
    *)
      log ""
      log "NOTE: ${BIN_DIR} is not in your PATH."
      log "Add it with: export PATH=\"${BIN_DIR}:\$PATH\""
      ;;
  esac
}

# --- Main ---

main() {
  need_cmd curl
  need_cmd uname

  detect_platform
  resolve_version
  download_and_verify

  case "$PLATFORM" in
    darwin) install_darwin ;;
    linux)  install_linux ;;
  esac

  log ""
  log "Done! Panen ${VERSION} installed successfully."
}

main
