#!/usr/bin/env bash

set -euo pipefail

# Directory containing the .txt files
SCRIPT_DIR="$(dirname "$(realpath "${BASH_SOURCE[0]}")")"

# Loop through each .txt file in the directory
for file in "$SCRIPT_DIR"/*.txt; do
  if [[ -f "$file" ]]; then
    echo "Sorting and deduping $file"
    sort -u "$file" -o "$file"
  fi
done

echo "Done."
