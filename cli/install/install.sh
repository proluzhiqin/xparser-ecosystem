#!/bin/sh
set -e

# xParser CLI installer
# Usage: curl -fsSL https://dllf.intsig.net/download/2026/Solution/xparser/install.sh | sh
#
# Environment variables:
#   XPARSER_VERSION   - version to install (default: "latest")
#   XPARSER_BASE_URL  - override download base URL
#   INSTALL_DIR       - install directory (default: ~/.local/bin)

VERSION="${XPARSER_VERSION:-latest}"
BASE_URL="${XPARSER_BASE_URL:-https://dllf.intsig.net/download/2026/Solution/xparser}"
INSTALL_DIR="${INSTALL_DIR:-$HOME/.local/bin}"

detect_platform() {
    OS="$(uname -s)"
    ARCH="$(uname -m)"

    case "$OS" in
        Linux)  OS="linux" ;;
        Darwin) OS="darwin" ;;
        *)      echo "Error: unsupported OS: $OS"; exit 1 ;;
    esac

    case "$ARCH" in
        x86_64|amd64)   ARCH="amd64" ;;
        aarch64|arm64)  ARCH="arm64" ;;
        *)              echo "Error: unsupported architecture: $ARCH"; exit 1 ;;
    esac
}

ensure_dir() {
    if [ -d "$INSTALL_DIR" ] && [ -w "$INSTALL_DIR" ]; then
        return
    fi

    # Try creating as current user first
    if mkdir -p "$INSTALL_DIR" 2>/dev/null; then
        return
    fi

    # Fallback to sudo
    if command -v sudo >/dev/null 2>&1; then
        echo "Elevated permissions required for ${INSTALL_DIR}"
        sudo mkdir -p "$INSTALL_DIR"
        sudo chown "$(id -u):$(id -g)" "$INSTALL_DIR"
    else
        echo "Error: cannot create ${INSTALL_DIR} and sudo not available"
        echo "Try: INSTALL_DIR=~/.local/bin sh install.sh"
        exit 1
    fi
}

ensure_path() {
    # Already in PATH — nothing to do
    case ":$PATH:" in
        *":${INSTALL_DIR}:"*) return ;;
    esac

    # Detect login shell (not the shell running this script)
    SHELL_NAME="$(basename "${SHELL:-sh}")"
    case "$SHELL_NAME" in
        zsh)  PROFILE="$HOME/.zshrc" ;;
        bash) PROFILE="${HOME}/.bashrc" ;;
        fish) PROFILE="$HOME/.config/fish/config.fish" ;;
        *)    PROFILE="$HOME/.profile" ;;
    esac

    # Already configured in profile
    if grep -qF "$INSTALL_DIR" "$PROFILE" 2>/dev/null; then
        return
    fi

    # Build export line
    if [ "$SHELL_NAME" = "fish" ]; then
        EXPORT_LINE="set -gx PATH ${INSTALL_DIR} \$PATH"
    else
        EXPORT_LINE="export PATH=\"${INSTALL_DIR}:\$PATH\""
    fi

    # Ensure parent dir exists (for fish config path)
    mkdir -p "$(dirname "$PROFILE")" 2>/dev/null || true

    printf '\n# Added by xparser installer\n%s\n' "$EXPORT_LINE" >> "$PROFILE"
    echo "Added ${INSTALL_DIR} to PATH in ${PROFILE}"
    echo "Run: source ${PROFILE}  (or restart your terminal)"
}

download_and_install() {
    BINARY="xparser-${OS}-${ARCH}"
    URL="${BASE_URL}/${VERSION}/${BINARY}"
    TMP="$(mktemp)"
    trap 'rm -f "$TMP"' EXIT

    echo "Downloading xparser ${VERSION} for ${OS}/${ARCH}..."
    echo "  ${URL}"

    if command -v curl >/dev/null 2>&1; then
        curl -fsSL -o "$TMP" "$URL"
    elif command -v wget >/dev/null 2>&1; then
        wget -qO "$TMP" "$URL"
    else
        echo "Error: curl or wget required"
        exit 1
    fi

    if [ ! -s "$TMP" ]; then
        echo "Error: download failed or file is empty"
        rm -f "$TMP"
        exit 1
    fi

    ensure_dir

    mv "$TMP" "${INSTALL_DIR}/xparser"
    chmod +x "${INSTALL_DIR}/xparser"

    ensure_path

    echo ""
    echo "Installed successfully!"
    echo "=========================================================="
    "${INSTALL_DIR}/xparser" version
    echo "=========================================================="
    echo ""
    echo "Executable: ${INSTALL_DIR}/xparser"
}

detect_platform
download_and_install
