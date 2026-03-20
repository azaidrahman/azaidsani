#!/bin/bash

set -e

if [ -z "$1" ]; then
  echo "Usage: ./scripts/add-image.sh <image-path> [image-path2] ..."
  echo "Example: ./scripts/add-image.sh media/My\ Photo.png media/another.webp"
  exit 1
fi

if ! command -v sips &>/dev/null; then
  echo "Error: sips not found. This script requires macOS."
  exit 1
fi

mkdir -p static/images
shortcodes=()

for src in "$@"; do
  if [ ! -f "$src" ]; then
    echo "Error: file not found: $src (skipping)"
    continue
  fi

  # Get extension
  ext="${src##*.}"
  ext=$(echo "$ext" | tr '[:upper:]' '[:lower:]')

  # Clean filename: lowercase, strip parens/brackets, spaces to hyphens, collapse hyphens
  basename="${src##*/}"
  name="${basename%.*}"
  clean=$(echo "$name" | tr '[:upper:]' '[:lower:]' | sed 's/[()[\]]//g' | sed 's/ \+/-/g' | sed 's/-\+/-/g' | sed 's/^-//;s/-$//')

  cleaned="${clean}.${ext}"

  echo "--- $basename ---"
  echo "Cleaned filename: $cleaned"

  # Ask for caption
  printf "Caption (optional): "
  read -r caption

  # Move
  mv "$src" "static/images/${cleaned}"
  echo "Moved to: static/images/${cleaned}"
  echo ""

  # Detect dimensions and pick tag
  dims=$(sips -g pixelWidth -g pixelHeight "static/images/${cleaned}")
  width=$(echo "$dims" | awk '/pixelWidth/{print $2}')
  height=$(echo "$dims" | awk '/pixelHeight/{print $2}')

  if [ "$((width))" -gt "$((height * 3 / 2))" ]; then
    tag="movies"
  else
    tag="mid-img"
  fi
  echo "Detected: ${width}x${height} -> ${tag}"

  # Collect shortcode
  if [ -n "$caption" ]; then
    shortcodes+=("{{< ${tag} src=\"/images/${cleaned}\" caption=\"${caption}\" >}}")
  else
    shortcodes+=("{{< ${tag} src=\"/images/${cleaned}\" >}}")
  fi
done

# Print all shortcodes at the end
if [ ${#shortcodes[@]} -gt 0 ]; then
  echo "========================="
  echo "Paste these into your post:"
  echo ""
  all=""
  for sc in "${shortcodes[@]}"; do
    echo "$sc"
    echo ""
    all+="$sc"$'\n'
  done
  printf "%s" "$all" | pbcopy
  echo "(Copied to clipboard)"
fi
