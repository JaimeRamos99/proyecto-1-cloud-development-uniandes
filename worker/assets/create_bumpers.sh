#!/bin/sh
# Script to create ANB intro and outro bumpers

ASSETS_DIR="/app/assets"
WATERMARK_FILE="$ASSETS_DIR/watermark.png"
INTRO_FILE="$ASSETS_DIR/intro.mp4"
OUTRO_FILE="$ASSETS_DIR/outro.mp4"

# Check if FFmpeg is available
if ! command -v ffmpeg &> /dev/null; then
    echo "Error: FFmpeg not found"
    exit 1
fi

# Check if watermark exists
if [ ! -f "$WATERMARK_FILE" ]; then
    echo "Error: watermark.png not found at $WATERMARK_FILE"
    exit 1
fi

echo "Creating ANB bumpers..."

# Create intro bumper (2.5 seconds)
# Black background with centered ANB logo, fade in effect
ffmpeg -f lavfi -i color=black:size=1280x720:duration=2.5:rate=30 \
    -i "$WATERMARK_FILE" \
    -filter_complex "[1:v]scale=400:160:force_original_aspect_ratio=decrease[logo];[0:v][logo]overlay=(main_w-overlay_w)/2:(main_h-overlay_h)/2,fade=in:0:15" \
    -c:v libx264 -pix_fmt yuv420p -crf 23 \
    -y "$INTRO_FILE" || {
    echo "Error: Failed to create intro bumper"
    exit 1
}

# Create outro bumper (2.5 seconds)  
# Black background with centered ANB logo, fade out effect
ffmpeg -f lavfi -i color=black:size=1280x720:duration=2.5:rate=30 \
    -i "$WATERMARK_FILE" \
    -filter_complex "[1:v]scale=400:160:force_original_aspect_ratio=decrease[logo];[0:v][logo]overlay=(main_w-overlay_w)/2:(main_h-overlay_h)/2,fade=out:60:15" \
    -c:v libx264 -pix_fmt yuv420p -crf 23 \
    -y "$OUTRO_FILE" || {
    echo "Error: Failed to create outro bumper"  
    exit 1
}

echo "✅ ANB bumpers created successfully:"
echo "   - Intro: $INTRO_FILE (2.5s)"
echo "   - Outro: $OUTRO_FILE (2.5s)"
echo "   - Logo size: 400x160px (centered)"
echo "   - Effects: Fade in/out"
