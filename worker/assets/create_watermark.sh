#!/bin/bash
# Script to create ANB watermark using ImageMagick

WATERMARK_FILE="/app/assets/watermark.png"

# Check if ImageMagick is available
if ! command -v convert &> /dev/null; then
    echo "Warning: ImageMagick not found. Creating simple placeholder watermark."
    echo "Install ImageMagick in Dockerfile with: RUN apk add --no-cache imagemagick"
    exit 1
fi

# Create ANB watermark with ImageMagick
convert -size 200x50 xc:none \
    -font Arial-Bold -pointsize 24 \
    -fill 'rgba(255,255,255,0.8)' \
    -stroke 'rgba(0,0,0,0.5)' -strokewidth 1 \
    -gravity center -annotate 0 "ANB" \
    "$WATERMARK_FILE"

echo "ANB watermark created at: $WATERMARK_FILE"
