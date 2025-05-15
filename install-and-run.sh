#!/bin/bash

REPO_URL="https://github.com/ascerra/tree-ai"
REPO_NAME="tree-ai"
TMP_DIR="tree-ai-tmp"

# Handle first argument as path to run on
TARGET_PATH="$1"
if [ -z "$TARGET_PATH" ]; then
  TARGET_PATH="."
fi

# Clone the repo if not already present
if [ ! -d "$TMP_DIR" ]; then
  echo "📦 Cloning tree-ai repo..."
  git clone --depth=1 "$REPO_URL" "$TMP_DIR" || {
    echo "❌ Failed to clone repo." >&2
    exit 1
  }
fi

cd "$TMP_DIR"

# Ensure Go dependencies are resolved
echo "📦 Running go mod tidy..."
go mod tidy

# Install (build Go + set up Python venv)
echo "⚙️  Installing tree-ai..."
make install || {
  echo "❌ Failed to install tree-ai. Make sure Go and Python 3 are available." >&2
  exit 1
}

# Run tree-ai
#./bin/tree-ai "$TARGET_PATH"

# Optionally clean up (comment out to persist install)
# echo "🧹 Cleaning up..."
# cd .. && rm -rf "$TMP_DIR"
