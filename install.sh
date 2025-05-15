# TODO

# #!/bin/bash

# set -e

# REPO_URL="https://github.com/ascerra/tree-ai"
# INSTALL_DIR="$HOME/.local/bin"
# REPO_NAME="tree-ai"
# TMP_DIR="$(mktemp -d)"

# echo "📥 Cloning latest tree-ai into $TMP_DIR"
# git clone --depth=1 "$REPO_URL" "$TMP_DIR"

# cd "$TMP_DIR"

# echo "📦 Ensuring Go dependencies..."
# go mod tidy

# echo "🔨 Building tree-ai binary..."
# go build -o "$INSTALL_DIR/tree-ai" ./main.go

# echo "⚙️ Setting up local Python model runner..."
# python3 -m venv .venv
# source .venv/bin/activate
# pip install --upgrade pip torch transformers
# go build -o "$INSTALL_DIR/granite-runner" model/granite-runner.go

# echo "🧹 Cleaning up temporary files..."
# deactivate
# rm -rf "$TMP_DIR"

# echo "✅ Installed tree-ai to $INSTALL_DIR"

# if ! echo "$PATH" | grep -q "$INSTALL_DIR"; then
#   echo ""
#   echo "📌 Please add the following to your shell config (~/.bashrc or ~/.zshrc):"
#   echo "    export PATH=\"\$PATH:$INSTALL_DIR\""
# fi

# echo ""
# echo "👉 You can now run:"
# echo "    tree-ai ./"
