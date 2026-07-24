#!/bin/sh
set -eu

# piomx installer — downloads the latest release binary from GitHub.
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/runtime-terror404/piomx/main/install.sh | sh
#   ./install.sh                # install or update
#   ./install.sh --uninstall    # remove binary + optional config cleanup

REPO="runtime-terror404/piomx"
BIN_NAME="piomx"

# ---- helpers ----

detect_platform() {
    case "$(uname -s)" in
        Linux)  OS="linux" ;;
        Darwin) OS="darwin" ;;
        *)      echo "Error: unsupported OS: $(uname -s)" >&2; exit 1 ;;
    esac

    case "$(uname -m)" in
        x86_64|amd64)  ARCH="amd64" ;;
        aarch64|arm64) ARCH="arm64" ;;
        *)             echo "Error: unsupported architecture: $(uname -m)" >&2; exit 1 ;;
    esac
}

install_dir() {
    if [ -n "${XDG_BIN_HOME:-}" ]; then
        echo "$XDG_BIN_HOME"
    elif [ -d "$HOME/.local/bin" ]; then
        echo "$HOME/.local/bin"
    else
        mkdir -p "$HOME/.local/bin"
        echo "$HOME/.local/bin"
    fi
}

# ---- uninstall ----

do_uninstall() {
    INSTALL_DIR=$(install_dir)
    BIN_PATH="$INSTALL_DIR/$BIN_NAME"

    if [ -f "$BIN_PATH" ]; then
        echo "Removing $BIN_PATH"
        rm -f "$BIN_PATH"
        echo "piomx uninstalled."
    else
        echo "piomx is not installed at $BIN_PATH"
    fi

    CONFIG_DIR="${XDG_CONFIG_HOME:-$HOME/.config}/piomx"
    if [ -d "$CONFIG_DIR" ]; then
        printf "Remove config at %s? [y/N]: " "$CONFIG_DIR"
        read -r answer
        case "$answer" in
            y|Y|yes) rm -rf "$CONFIG_DIR" && echo "Config removed." ;;
            *) echo "Config kept at $CONFIG_DIR" ;;
        esac
    fi
    exit 0
}

# ---- install ----

do_install() {
    detect_platform
    INSTALL_DIR=$(install_dir)
    BIN_PATH="$INSTALL_DIR/$BIN_NAME"
    TMP_DIR=$(mktemp -d)
    trap 'rm -rf "$TMP_DIR"' EXIT

    # Fetch latest release URL from GitHub API.
    echo "Fetching latest release for $OS/$ARCH..."
    RELEASE_URL=$(curl -fsSL "https://api.github.com/repos/$REPO/releases/latest" \
        | grep "browser_download_url" \
        | grep "piomx_${OS}_${ARCH}" \
        | head -1 \
        | cut -d '"' -f 4)

    if [ -z "$RELEASE_URL" ]; then
        echo "Error: could not find release for $OS/$ARCH" >&2
        exit 1
    fi

    echo "Downloading $RELEASE_URL ..."
    curl -fsSL "$RELEASE_URL" -o "$TMP_DIR/$BIN_NAME.tar.gz"

	tar -xzf "$TMP_DIR/$BIN_NAME.tar.gz" -C "$TMP_DIR"
    chmod +x "$TMP_DIR/$BIN_NAME"

    if [ -f "$BIN_PATH" ]; then
        echo "Replacing existing installation at $BIN_PATH"
    fi

    mv "$TMP_DIR/$BIN_NAME" "$BIN_PATH"
    echo "Installed $BIN_NAME to $BIN_PATH"

    # Check PATH.
    if ! echo "$PATH" | tr ':' '\n' | grep -Fxq "$INSTALL_DIR"; then
        echo ""
        echo "Note: $INSTALL_DIR is not in your PATH."
        echo "Add this to your shell config:"
        echo "  export PATH=\"\$PATH:$INSTALL_DIR\""
    fi

    echo ""
    echo "Done. Run 'piomx' to get started."
}

# ---- main ----

case "${1:-}" in
    --uninstall|-u) do_uninstall ;;
    *) do_install ;;
esac
